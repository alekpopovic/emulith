package dynamodb

import "testing"

func TestParseUpdate(t *testing.T) {
	a, e := parseUpdate("SET #a = :x, #b=:y REMOVE #c")
	if e != nil || len(a) != 3 || a[0].kind != "SET" || a[2].kind != "REMOVE" {
		t.Fatalf("%#v %v", a, e)
	}
	for _, s := range []string{"", "#a=:x", "SET #a"} {
		if _, e := parseUpdate(s); e == nil {
			t.Fatalf("accepted %q", s)
		}
	}
}
