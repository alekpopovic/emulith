package blob

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestParseRange(t *testing.T) {
	for in, want := range map[string]Range{"bytes=2-4": {2, 4, 3}, "bytes=2-": {2, 9, 8}, "bytes=-3": {7, 9, 3}} {
		r, e := ParseRange(in, 10)
		if e != nil || r != want {
			t.Errorf("%s: %#v %v", in, r, e)
		}
	}
	if _, e := ParseRange("bytes=10-", 10); e == nil {
		t.Fatal("expected invalid")
	}
}
func TestCommitBlocks(t *testing.T) {
	a := map[string][]byte{}
	i, _ := CanonicalBlockID("YQ==")
	a[i] = []byte("a")
	j, _ := CanonicalBlockID("Yg==")
	a[j] = []byte("b")
	b, e := CommitBlocks(a, []string{"Yg==", "YQ=="}, 10)
	if e != nil || string(b) != "ba" {
		t.Fatal(string(b), e)
	}
}
func TestConditions(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("If-None-Match", "\"x\"")
	if Conditions(r, "\"x\"", time.Now()) != 304 {
		t.Fatal()
	}
	r.Header.Set("If-Match", "\"y\"")
	if Conditions(r, "\"x\"", time.Now()) != 412 {
		t.Fatal()
	}
	_ = strings.Builder{}
}
