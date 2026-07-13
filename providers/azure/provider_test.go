package azure

import (
	"context"
	"github.com/alekpopovic/emulith/internal/state"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestTableLifecycleAndQuery(t *testing.T) {
	d, _ := os.MkdirTemp("", "az")
	defer os.RemoveAll(d)
	s, e := state.Open(context.Background(), d)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	h := NewWithStore("table", "devstoreaccount1", s)
	req := httptest.NewRequest("POST", "/devstoreaccount1/Users", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Fatalf("create table: %d", w.Code)
	}
	body := strings.NewReader(`{"PartitionKey":"p","RowKey":"r","name":"alice"}`)
	req = httptest.NewRequest("POST", "/devstoreaccount1/Users/p/r", body)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code < 200 || w.Code >= 300 {
		t.Fatalf("insert: %d", w.Code)
	}
	req = httptest.NewRequest("GET", "/devstoreaccount1/Users?$filter=PartitionKey%20eq%20'p'", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("query: %d", w.Code)
	}
}
