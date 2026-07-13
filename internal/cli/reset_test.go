package cli

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResetCommand(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_emulith/reset" || r.Method != "POST" {
			t.Fatalf("request %s %s", r.Method, r.URL.Path)
		}
		io.WriteString(w, `{"status":"ok","reset":true}`)
	}))
	defer srv.Close()
	var out bytes.Buffer
	cmd := NewCommandWithClient(&out, &bytes.Buffer{}, "dev", "unknown", "unknown", srv.Client())
	cmd.SetArgs([]string{"reset", "--endpoint", srv.URL})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if out.Len() == 0 {
		t.Fatal("missing output")
	}
}
func TestResetMalformed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "bad") }))
	defer srv.Close()
	cmd := NewCommandWithClient(&bytes.Buffer{}, &bytes.Buffer{}, "dev", "unknown", "unknown", srv.Client())
	cmd.SetArgs([]string{"reset", "--endpoint", srv.URL})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}
