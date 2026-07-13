package dynamodb

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

type updateRHS struct {
	kind        string
	path        []pathPart
	value       Value
	left, right *updateRHS
}
type plannedAction struct {
	kind  string
	path  []pathPart
	rhs   *updateRHS
	value Value
}
type updatePlan struct{ actions []plannedAction }
type updateParser struct {
	expressionParser
	sections map[string]bool
}

func parseUpdatePlan(input string, names map[string]string, values map[string]Value) (updatePlan, error) {
	tokens, e := lexExpression(input)
	if e != nil {
		return updatePlan{}, e
	}
	p := &updateParser{expressionParser: expressionParser{tokens: tokens, names: names, values: values}, sections: map[string]bool{}}
	var plan updatePlan
	for p.peek().kind != tEOF {
		if p.peek().kind != tWord {
			return plan, fmt.Errorf("expected update section at position %d", p.peek().pos)
		}
		section := strings.ToUpper(p.take().text)
		if section != "SET" && section != "REMOVE" && section != "ADD" && section != "DELETE" {
			return plan, fmt.Errorf("unsupported update section %s", section)
		}
		if p.sections[section] {
			return plan, fmt.Errorf("duplicate %s section", section)
		}
		p.sections[section] = true
		for {
			path, e := p.parsePath()
			if e != nil {
				return plan, e
			}
			a := plannedAction{kind: section, path: path}
			if section == "SET" {
				if p.peek().kind != tEq {
					return plan, errors.New("SET requires =")
				}
				p.take()
				a.rhs, e = p.parseRHS()
				if e != nil {
					return plan, e
				}
			} else if section == "ADD" || section == "DELETE" {
				if len(path) != 1 {
					return plan, fmt.Errorf("%s only supports top-level attributes", section)
				}
				if p.peek().kind != tValue {
					return plan, fmt.Errorf("%s requires a value", section)
				}
				tok := p.take()
				var ok bool
				a.value, ok = values[tok.text]
				if !ok {
					return plan, fmt.Errorf("missing value placeholder %s", tok.text)
				}
			}
			for _, old := range plan.actions {
				if pathsOverlap(old.path, a.path) {
					return plan, errors.New("update paths overlap")
				}
			}
			plan.actions = append(plan.actions, a)
			if p.peek().kind == tComma {
				p.take()
				continue
			}
			break
		}
	}
	if len(plan.actions) == 0 {
		return plan, errors.New("empty update expression")
	}
	return plan, nil
}
func (p *updateParser) parseRHS() (*updateRHS, error) {
	left, e := p.parseTerm()
	if e != nil {
		return nil, e
	}
	if p.peek().kind == tPlus || p.peek().kind == tMinus {
		op := p.take()
		right, e := p.parseTerm()
		if e != nil {
			return nil, e
		}
		kind := "+"
		if op.kind == tMinus {
			kind = "-"
		}
		return &updateRHS{kind: kind, left: left, right: right}, nil
	}
	return left, nil
}
func (p *updateParser) parseTerm() (*updateRHS, error) {
	if p.peek().kind == tValue {
		tok := p.take()
		v, ok := p.values[tok.text]
		if !ok {
			return nil, fmt.Errorf("missing value placeholder %s", tok.text)
		}
		return &updateRHS{kind: "value", value: v}, nil
	}
	if p.peek().kind == tWord && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].kind == tLParen {
		name := strings.ToLower(p.take().text)
		if name != "if_not_exists" && name != "list_append" {
			return nil, fmt.Errorf("unsupported SET function %s", name)
		}
		p.take()
		a, e := p.parseRHS()
		if e != nil {
			return nil, e
		}
		if p.peek().kind != tComma {
			return nil, errors.New("function requires two arguments")
		}
		p.take()
		b, e := p.parseRHS()
		if e != nil {
			return nil, e
		}
		if p.peek().kind != tRParen {
			return nil, errors.New("function missing )")
		}
		p.take()
		return &updateRHS{kind: name, left: a, right: b}, nil
	}
	path, e := p.parsePath()
	return &updateRHS{kind: "path", path: path}, e
}
func pathsOverlap(a, b []pathPart) bool {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func applyUpdatePlan(plan updatePlan, item map[string]Value, partition, sortKey string) error {
	for _, a := range plan.actions {
		if len(a.path) == 0 || a.path[0].name == partition || a.path[0].name == sortKey {
			return errors.New("cannot update primary key")
		}
		switch a.kind {
		case "SET":
			v, e := evalUpdateRHS(a.rhs, item)
			if e != nil {
				return e
			}
			if e = setPath(item, a.path, v); e != nil {
				return e
			}
		case "REMOVE":
			if e := removePath(item, a.path); e != nil {
				return e
			}
		case "ADD":
			if e := addPath(item, a.path, a.value); e != nil {
				return e
			}
		case "DELETE":
			if e := deleteFromSet(item, a.path, a.value); e != nil {
				return e
			}
		}
	}
	return nil
}
func evalUpdateRHS(r *updateRHS, item map[string]Value) (Value, error) {
	switch r.kind {
	case "value":
		return r.value, nil
	case "path":
		v, ok, e := getPath(item, r.path)
		if e != nil {
			return Value{}, e
		}
		if !ok {
			return Value{}, errors.New("referenced update path does not exist")
		}
		return v, nil
	case "+", "-":
		a, e := evalUpdateRHS(r.left, item)
		if e != nil {
			return Value{}, e
		}
		b, e := evalUpdateRHS(r.right, item)
		if e != nil {
			return Value{}, e
		}
		if a.Kind != "N" || b.Kind != "N" {
			return Value{}, errors.New("arithmetic requires numbers")
		}
		x, _ := new(big.Rat).SetString(a.N)
		y, _ := new(big.Rat).SetString(b.N)
		if r.kind == "+" {
			x.Add(x, y)
		} else {
			x.Sub(x, y)
		}
		return Value{Kind: "N", N: ratDecimal(x)}, nil
	case "if_not_exists":
		if r.left.kind != "path" {
			return Value{}, errors.New("if_not_exists first argument must be path")
		}
		if v, ok, e := getPath(item, r.left.path); e != nil {
			return Value{}, e
		} else if ok {
			return v, nil
		}
		return evalUpdateRHS(r.right, item)
	case "list_append":
		a, e := evalUpdateRHS(r.left, item)
		if e != nil {
			return Value{}, e
		}
		b, e := evalUpdateRHS(r.right, item)
		if e != nil {
			return Value{}, e
		}
		if a.Kind != "L" || b.Kind != "L" {
			return Value{}, errors.New("list_append requires lists")
		}
		return Value{Kind: "L", L: append(append([]Value{}, a.L...), b.L...)}, nil
	}
	return Value{}, errors.New("unsupported update value")
}
func ratDecimal(r *big.Rat) string {
	if r.IsInt() {
		return r.Num().String()
	}
	s := r.FloatString(38)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	if s == "-0" {
		return "0"
	}
	return s
}
func setPath(item map[string]Value, path []pathPart, v Value) error {
	if len(path) == 1 {
		item[path[0].name] = v
		return nil
	}
	root, ok := item[path[0].name]
	if !ok {
		return errors.New("parent update path does not exist")
	}
	if e := setNested(&root, path[1:], v, false); e != nil {
		return e
	}
	item[path[0].name] = root
	return nil
}
func removePath(item map[string]Value, path []pathPart) error {
	if len(path) == 1 {
		delete(item, path[0].name)
		return nil
	}
	root, ok := item[path[0].name]
	if !ok {
		return nil
	}
	if e := setNested(&root, path[1:], Value{}, true); e != nil {
		return e
	}
	item[path[0].name] = root
	return nil
}
func setNested(cur *Value, path []pathPart, v Value, remove bool) error {
	p := path[0]
	if p.isIdx {
		if cur.Kind != "L" || p.index >= len(cur.L) {
			return errors.New("list index out of range")
		}
		if len(path) == 1 {
			if remove {
				cur.L = append(cur.L[:p.index], cur.L[p.index+1:]...)
			} else {
				cur.L[p.index] = v
			}
			return nil
		}
		child := cur.L[p.index]
		if e := setNested(&child, path[1:], v, remove); e != nil {
			return e
		}
		cur.L[p.index] = child
		return nil
	}
	if cur.Kind != "M" {
		return errors.New("document path parent is not a map")
	}
	if len(path) == 1 {
		if remove {
			delete(cur.M, p.name)
		} else {
			cur.M[p.name] = v
		}
		return nil
	}
	child, ok := cur.M[p.name]
	if !ok {
		return errors.New("parent update path does not exist")
	}
	if e := setNested(&child, path[1:], v, remove); e != nil {
		return e
	}
	cur.M[p.name] = child
	return nil
}
func addPath(item map[string]Value, path []pathPart, v Value) error {
	old, ok, e := getPath(item, path)
	if e != nil {
		return e
	}
	if !ok {
		return setPath(item, path, v)
	}
	if old.Kind == "N" && v.Kind == "N" {
		x, _ := new(big.Rat).SetString(old.N)
		y, _ := new(big.Rat).SetString(v.N)
		x.Add(x, y)
		return setPath(item, path, Value{Kind: "N", N: ratDecimal(x)})
	}
	if old.Kind != v.Kind || (old.Kind != "SS" && old.Kind != "NS" && old.Kind != "BS") {
		return errors.New("ADD requires matching numbers or sets")
	}
	merged := append(append([]Value{}, old.Set...), v.Set...)
	seen := map[string]bool{}
	out := []Value{}
	for _, x := range merged {
		b, _ := x.Canonical()
		if !seen[string(b)] {
			seen[string(b)] = true
			out = append(out, x)
		}
	}
	old.Set = out
	return setPath(item, path, old)
}
func deleteFromSet(item map[string]Value, path []pathPart, v Value) error {
	old, ok, e := getPath(item, path)
	if e != nil || !ok {
		return e
	}
	if old.Kind != v.Kind || (old.Kind != "SS" && old.Kind != "NS" && old.Kind != "BS") {
		return errors.New("DELETE requires matching sets")
	}
	remove := map[string]bool{}
	for _, x := range v.Set {
		b, _ := x.Canonical()
		remove[string(b)] = true
	}
	out := []Value{}
	for _, x := range old.Set {
		b, _ := x.Canonical()
		if !remove[string(b)] {
			out = append(out, x)
		}
	}
	if len(out) == 0 {
		return removePath(item, path)
	}
	old.Set = out
	return setPath(item, path, old)
}
