package aws

import (
	"github.com/alekpopovic/emulith/internal/service"
	"io"
	"log/slog"
	"testing"
)

func TestProviderRegistration(t *testing.T) {
	p := NewProvider(nil, slog.New(slog.NewTextHandler(io.Discard, nil)), service.NewRegistry())
	if err := p.Register("custom", placeholder{}); err != nil {
		t.Fatal(err)
	}
	if err := p.Register("custom", placeholder{}); err == nil {
		t.Fatal("duplicate accepted")
	}
}
