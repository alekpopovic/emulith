package common

import "testing"

func TestProjectAndResourceValidation(t *testing.T) {
	for _, p := range []string{"emulith-local", "abc123"} {
		if !ValidProject(p) {
			t.Fatalf("valid project rejected: %s", p)
		}
	}
	for _, p := range []string{"", "A", "ab", "bad_"} {
		if ValidProject(p) {
			t.Fatalf("invalid project accepted: %s", p)
		}
	}
	r, e := Resource("emulith-local", "topics", "events")
	if e != nil || r != "projects/emulith-local/topics/events" {
		t.Fatalf("resource: %s %v", r, e)
	}
}
