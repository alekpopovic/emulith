package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestAzureConnectionString(t *testing.T) {
	var out, err bytes.Buffer
	c := NewCommand(&out, &err, "dev", "c", "b")
	c.SetArgs([]string{"azure", "connection-string", "--host", "127.0.0.1"})
	if e := c.Execute(); e != nil {
		t.Fatal(e)
	}
	s := out.String()
	for _, want := range []string{"AccountName=devstoreaccount1", "BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1", "QueueEndpoint=http://127.0.0.1:10001/devstoreaccount1", "TableEndpoint=http://127.0.0.1:10002/devstoreaccount1"} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %s in %s", want, s)
		}
	}
}
