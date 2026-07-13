package server

import (
	"context"
	"github.com/alekpopovic/emulith/internal/state"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestResetHandler(t *testing.T) {
	s, err := state.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.CreateS3Bucket(context.Background(), state.S3Bucket{Name: "bucket", Region: "us-east-1", CreatedAt: time.Now()})
	h := NewWithState(":0", "dev", s, slog.New(slog.NewTextHandler(io.Discard, nil)), http.NotFoundHandler()).HTTPServer().Handler
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/_emulith/reset", nil))
	if w.Code != 405 {
		t.Fatalf("GET=%d", w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/_emulith/reset", nil))
	if w.Code != 200 {
		t.Fatalf("POST=%d %s", w.Code, w.Body.String())
	}
	b, _ := s.ListS3Buckets(context.Background())
	if len(b) != 0 {
		t.Fatalf("buckets remain")
	}
}
