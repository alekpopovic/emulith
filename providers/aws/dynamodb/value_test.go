package dynamodb

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestValuesAndCanonicalKeys(t *testing.T) {
	samples := []string{`{"S":"x"}`, `{"N":"12345678901234567890.001"}`, `{"B":"AAE="}`, `{"BOOL":true}`, `{"NULL":true}`, `{"M":{"x":{"L":[{"S":"y"}]}}}`, `{"SS":["b","a"]}`, `{"NS":["1","2"]}`, `{"BS":["AA==","AQ=="]}`}
	for _, s := range samples {
		var v Value
		if err := json.Unmarshal([]byte(s), &v); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
		b, e := json.Marshal(v)
		if e != nil || !json.Valid(b) {
			t.Fatal(string(b), e)
		}
	}
	for _, s := range []string{`{"S":"x","N":"1"}`, `{"N":"01"}`, `{"SS":["a","a"]}`, `{"SS":[]}`} {
		var v Value
		if json.Unmarshal([]byte(s), &v) == nil {
			t.Fatalf("accepted %s", s)
		}
	}
	schema := KeySchema{"pk", "N", "sk", "B"}
	a, e := EncodeKey(schema, map[string]Value{"pk": {Kind: "N", N: "1.0"}, "sk": {Kind: "B", B: []byte{0, 1}}})
	if e != nil {
		t.Fatal(e)
	}
	b, e := EncodeKey(schema, map[string]Value{"pk": {Kind: "N", N: "1"}, "sk": {Kind: "B", B: []byte{0, 1}}})
	if e != nil || !bytes.Equal(a, b) {
		t.Fatal(a, b, e)
	}
}
func TestDepthLimit(t *testing.T) {
	s := `{"S":"x"}`
	for i := 0; i < MaxDepth+2; i++ {
		s = `{"L":[` + s + `]}`
	}
	var v Value
	if json.Unmarshal([]byte(s), &v) == nil {
		t.Fatal("accepted deep value")
	}
}
func FuzzValueJSON(f *testing.F) {
	f.Add([]byte(`{"N":"1.0"}`))
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 1<<20 {
			return
		}
		var v Value
		_ = json.Unmarshal(b, &v)
	})
}
