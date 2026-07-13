package compatibility

import (
	"strings"
	"testing"
)

func valid() Catalog {
	return Catalog{Entries: []Entry{{Provider: "aws", Service: "sts", Operation: "Get", Status: "supported", Protocol: "query", TestID: "id", Notes: "ok", Since: "v"}}}
}
func TestValidate(t *testing.T) {
	if err := Validate(valid(), []Result{{"id", "pass"}}, true); err != nil {
		t.Fatal(err)
	}
	c := valid()
	c.Entries[0].Status = "bad"
	if Validate(c, nil, false) == nil {
		t.Fatal("invalid status")
	}
	c = valid()
	c.Entries = append(c.Entries, c.Entries[0])
	if Validate(c, nil, false) == nil {
		t.Fatal("duplicate")
	}
	c = valid()
	c.Entries[0].TestID = ""
	if Validate(c, nil, false) == nil {
		t.Fatal("missing id")
	}
	if Validate(valid(), nil, true) != nil {
		t.Fatal("nil results only validates catalog")
	}
	if Validate(valid(), []Result{}, true) == nil {
		t.Fatal("missing test")
	}
	if Validate(valid(), []Result{{"id", "pass"}, {"id", "pass"}}, true) == nil {
		t.Fatal("duplicate result")
	}
}
func TestMarkdownEscapingDeterministic(t *testing.T) {
	c := valid()
	c.Entries[0].Notes = "a|b"
	a := Markdown(c)
	if !strings.Contains(a, "a\\|b") || a != Markdown(c) {
		t.Fatal(a)
	}
}
func TestJSONSchema(t *testing.T) {
	b, e := JSON(Report{SchemaVersion: 1})
	if e != nil || !strings.Contains(string(b), `"schema_version": 1`) {
		t.Fatal(string(b), e)
	}
}
