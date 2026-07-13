package dynamodb

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"sort"
)

const MaxDepth = 32
const MaxItemSize = 400 << 10
const MaxElements = 1000

var numberRE = regexp.MustCompile(`^-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?$`)

type Value struct {
	Kind string
	S    string
	N    string
	B    []byte
	Bool bool
	Null bool
	M    map[string]Value
	L    []Value
	Set  []Value
}

func (v *Value) UnmarshalJSON(data []byte) error { return v.decode(data, 0) }
func (v *Value) decode(data []byte, depth int) error {
	if depth > MaxDepth {
		return errors.New("attribute value nesting too deep")
	}
	var raw map[string]json.RawMessage
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&raw); err != nil {
		return err
	}
	if len(raw) != 1 {
		return errors.New("attribute value must contain exactly one variant")
	}
	for k, b := range raw {
		v.Kind = k
		switch k {
		case "S":
			return json.Unmarshal(b, &v.S)
		case "N":
			if err := json.Unmarshal(b, &v.N); err != nil {
				return err
			}
			return validateNumber(v.N)
		case "B":
			var s string
			if err := json.Unmarshal(b, &s); err != nil {
				return err
			}
			x, e := base64.StdEncoding.DecodeString(s)
			v.B = x
			return e
		case "BOOL":
			return json.Unmarshal(b, &v.Bool)
		case "NULL":
			if err := json.Unmarshal(b, &v.Null); err != nil {
				return err
			}
			if !v.Null {
				return errors.New("NULL must be true")
			}
			return nil
		case "M":
			var m map[string]json.RawMessage
			if err := json.Unmarshal(b, &m); err != nil {
				return err
			}
			if len(m) > MaxElements {
				return errors.New("map too large")
			}
			v.M = map[string]Value{}
			for n, r := range m {
				var x Value
				if err := x.decode(r, depth+1); err != nil {
					return fmt.Errorf("%s: %w", n, err)
				}
				v.M[n] = x
			}
			return nil
		case "L":
			var a []json.RawMessage
			if err := json.Unmarshal(b, &a); err != nil {
				return err
			}
			if len(a) > MaxElements {
				return errors.New("list too large")
			}
			for _, r := range a {
				var x Value
				if err := x.decode(r, depth+1); err != nil {
					return err
				}
				v.L = append(v.L, x)
			}
			return nil
		case "SS", "NS", "BS":
			return v.decodeSet(b, k)
		default:
			return fmt.Errorf("unknown attribute value variant %q", k)
		}
	}
	return nil
}
func (v *Value) decodeSet(b []byte, k string) error {
	var a []string
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	if len(a) == 0 || len(a) > MaxElements {
		return errors.New("set must be non-empty and bounded")
	}
	seen := map[string]bool{}
	for _, s := range a {
		var x Value
		switch k {
		case "SS":
			x = Value{Kind: "S", S: s}
		case "NS":
			if err := validateNumber(s); err != nil {
				return err
			}
			x = Value{Kind: "N", N: s}
		case "BS":
			z, e := base64.StdEncoding.DecodeString(s)
			if e != nil {
				return e
			}
			x = Value{Kind: "B", B: z}
		}
		c, _ := x.Canonical()
		q := string(c)
		if seen[q] {
			return errors.New("duplicate set member")
		}
		seen[q] = true
		v.Set = append(v.Set, x)
	}
	sort.Slice(v.Set, func(i, j int) bool {
		a, _ := v.Set[i].Canonical()
		b, _ := v.Set[j].Canonical()
		return bytes.Compare(a, b) < 0
	})
	return nil
}
func validateNumber(s string) error {
	if !numberRE.MatchString(s) {
		return errors.New("invalid DynamoDB number")
	}
	if _, ok := new(big.Rat).SetString(s); !ok {
		return errors.New("invalid DynamoDB number")
	}
	return nil
}
func (v Value) MarshalJSON() ([]byte, error) {
	var x any
	switch v.Kind {
	case "S":
		x = v.S
	case "N":
		x = v.N
	case "B":
		x = base64.StdEncoding.EncodeToString(v.B)
	case "BOOL":
		x = v.Bool
	case "NULL":
		x = true
	case "M":
		x = v.M
	case "L":
		x = v.L
	case "SS", "NS", "BS":
		a := []string{}
		for _, e := range v.Set {
			if e.Kind == "B" {
				a = append(a, base64.StdEncoding.EncodeToString(e.B))
			} else if e.Kind == "N" {
				a = append(a, e.N)
			} else {
				a = append(a, e.S)
			}
		}
		x = a
	default:
		return nil, errors.New("invalid value")
	}
	return json.Marshal(map[string]any{v.Kind: x})
}
func (v Value) Canonical() ([]byte, error) {
	if v.Kind == "N" {
		r, _ := new(big.Rat).SetString(v.N)
		return []byte("N\x00" + r.RatString()), nil
	}
	b, e := v.MarshalJSON()
	if e != nil {
		return nil, e
	}
	return append([]byte(v.Kind+"\x00"), b...), nil
}

type KeySchema struct {
	PartitionName string
	PartitionType string
	SortName      string
	SortType      string
}

func EncodeKey(schema KeySchema, item map[string]Value) ([]byte, error) {
	p, ok := item[schema.PartitionName]
	if !ok {
		return nil, errors.New("missing partition key")
	}
	a, e := keyPart(schema.PartitionType, p)
	if e != nil {
		return nil, e
	}
	out := frame(a)
	if schema.SortName != "" {
		s, ok := item[schema.SortName]
		if !ok {
			return nil, errors.New("missing sort key")
		}
		b, e := keyPart(schema.SortType, s)
		if e != nil {
			return nil, e
		}
		out = append(out, frame(b)...)
	}
	return out, nil
}
func keyPart(kind string, v Value) ([]byte, error) {
	if v.Kind != kind || (kind != "S" && kind != "N" && kind != "B") {
		return nil, errors.New("wrong key type")
	}
	return v.Canonical()
}
func frame(b []byte) []byte {
	x := make([]byte, 4+len(b))
	binary.BigEndian.PutUint32(x, uint32(len(b)))
	copy(x[4:], b)
	return x
}
