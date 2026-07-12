package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	var out bytes.Buffer
	cmd := NewCommand(&out, &bytes.Buffer{}, "1.2.3", "abc", "now")
	cmd.SetArgs([]string{"version"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	want := "emulith 1.2.3\ncommit: abc\nbuilt: now\n"
	if out.String() != want {
		t.Fatalf("output = %q", out.String())
	}
	if strings.Contains(out.String(), "unknown") {
		t.Fatal("unexpected default")
	}
}
