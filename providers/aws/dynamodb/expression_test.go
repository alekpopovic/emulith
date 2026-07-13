package dynamodb

import (
	"encoding/json"
	"strings"
	"testing"
)

func mustValue(t *testing.T, s string) Value {
	t.Helper()
	var v Value
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatal(err)
	}
	return v
}
func TestConditionPrecedenceFunctionsAndPaths(t *testing.T) {
	item := map[string]Value{"pk": {Kind: "S", S: "abc"}, "n": {Kind: "N", N: "10"}, "doc": {Kind: "M", M: map[string]Value{"list": {Kind: "L", L: []Value{{Kind: "S", S: "x"}, {Kind: "S", S: "y"}}}}}}
	names := map[string]string{"#p": "pk", "#n": "n", "#d": "doc"}
	values := map[string]Value{":a": {Kind: "S", S: "a"}, ":lo": {Kind: "N", N: "2"}, ":hi": {Kind: "N", N: "11"}, ":y": {Kind: "S", S: "y"}}
	tests := []string{"attribute_exists(#p) AND begins_with(#p, :a)", "#n BETWEEN :lo AND :hi", "contains(#d.list, :y)", "NOT attribute_not_exists(#p)", "size(#p) = :three"}
	values[":three"] = Value{Kind: "N", N: "3"}
	for _, s := range tests {
		e, err := parseCondition(s, names, values)
		if err != nil {
			t.Fatalf("%s: %v", s, err)
		}
		ok, err := evalCondition(e, item)
		if err != nil || !ok {
			t.Fatalf("%s => %v %v", s, ok, err)
		}
	}
}
func TestUpdatePlanArithmeticSetsNested(t *testing.T) {
	names := map[string]string{"#n": "n", "#x": "x", "#s": "set"}
	vals := map[string]Value{":one": {Kind: "N", N: "1.25"}, ":v": {Kind: "S", S: "new"}, ":ss": mustValue(t, `{"SS":["a","b"]}`)}
	p, e := parseUpdatePlan("SET #n = #n + :one, #x = :v ADD #s :ss", names, vals)
	if e != nil {
		t.Fatal(e)
	}
	item := map[string]Value{"pk": {Kind: "S", S: "k"}, "n": {Kind: "N", N: "2.75"}}
	if e = applyUpdatePlan(p, item, "pk", ""); e != nil {
		t.Fatal(e)
	}
	if item["n"].N != "4" || item["x"].S != "new" || len(item["set"].Set) != 2 {
		t.Fatalf("%#v", item)
	}
}
func TestExpressionLimitsAndMalformed(t *testing.T) {
	if _, e := lexExpression(strings.Repeat("x", maxExpressionLength+1)); e == nil {
		t.Fatal("accepted oversized")
	}
	for _, s := range []string{"a =", "(a = :v", "unknown(a)"} {
		if _, e := parseCondition(s, nil, map[string]Value{":v": {Kind: "S", S: "x"}}); e == nil {
			t.Fatalf("accepted %q", s)
		}
	}
}
func FuzzExpressionParser(f *testing.F) {
	f.Add("#a = :v")
	f.Fuzz(func(t *testing.T, s string) {
		if len(s) > maxExpressionLength+1 {
			return
		}
		_, _ = parseCondition(s, map[string]string{"#a": "a"}, map[string]Value{":v": {Kind: "S", S: "x"}})
	})
}
