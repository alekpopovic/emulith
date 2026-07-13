package storage

import (
	"context"
	"github.com/alekpopovic/emulith/internal/state"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestBucketLifecycle(t *testing.T) {
	d, _ := os.MkdirTemp("", "gcp")
	defer os.RemoveAll(d)
	s, e := state.Open(context.Background(), d)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	h := &Handler{Store: s, Project: "emulith-local"}
	r := httptest.NewRequest("POST", "/storage/v1/b?project=emulith-local", strings.NewReader(`{"name":"bucket-one"}`))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("insert %d", w.Code)
	}
	r = httptest.NewRequest("GET", "/storage/v1/b/bucket-one", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("get %d", w.Code)
	}
}
