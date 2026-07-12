package s3

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/emulith/emulith/internal/state"
	awsprovider "github.com/emulith/emulith/providers/aws"
)

func setup(t *testing.T) *Handler {
	t.Helper()
	s, err := state.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return New(s)
}
func call(t *testing.T, h *Handler, method, target, operation, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	req := &awsprovider.Request{HTTPRequest: r, Protocol: awsprovider.ProtocolS3, Service: "s3", Operation: operation}
	w := httptest.NewRecorder()
	h.ServeAWS(w, req, "id")
	return w
}
func TestBucketObjectLifecycle(t *testing.T) {
	h := setup(t)
	if w := call(t, h, "PUT", "/valid-bucket", "CreateBucket", ""); w.Code != 200 {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	key := "nested/ü space/../%value<&"
	targetKey := "nested/ü%20space/../%25value%3C%26"
	w := call(t, h, "PUT", "/valid-bucket/"+targetKey, "PutObject", "\x00binary")
	if w.Code != 200 {
		t.Fatalf("put: %d %s", w.Code, w.Body.String())
	}
	w = call(t, h, "GET", "/valid-bucket/"+targetKey, "GetObject", "")
	if w.Code != 200 || w.Body.String() != "\x00binary" {
		t.Fatalf("get: %d %q", w.Code, w.Body.String())
	}
	w = call(t, h, "HEAD", "/valid-bucket/"+targetKey, "HeadObject", "")
	if w.Code != 200 || w.Body.Len() != 0 {
		t.Fatalf("head: %d %q", w.Code, w.Body.String())
	}
	w = call(t, h, "GET", "/valid-bucket?list-type=2&prefix=nested&max-keys=10", "ListObjectsV2", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "&lt;&amp;") || !strings.Contains(w.Body.String(), "%value") {
		t.Fatalf("list key %q: %d %s", key, w.Code, w.Body.String())
	}
	w = call(t, h, "DELETE", "/valid-bucket/"+targetKey, "DeleteObject", "")
	if w.Code != 204 {
		t.Fatalf("delete: %d", w.Code)
	}
	w = call(t, h, "GET", "/valid-bucket/"+targetKey, "GetObject", "")
	if w.Code != 404 || !strings.Contains(w.Body.String(), "NoSuchKey") {
		t.Fatalf("missing: %d %s", w.Code, w.Body.String())
	}
}
func TestValidationAndOrdering(t *testing.T) {
	h := setup(t)
	for _, name := range []string{"Bad_Name", "ab", "192.168.1.1"} {
		if w := call(t, h, "PUT", "/"+name, "CreateBucket", ""); w.Code != 400 {
			t.Fatalf("accepted %q", name)
		}
	}
	for _, name := range []string{"z-bucket", "a-bucket"} {
		if w := call(t, h, "PUT", "/"+name, "CreateBucket", ""); w.Code != 200 {
			t.Fatal(w.Body.String())
		}
	}
	w := call(t, h, "GET", "/", "ListBuckets", "")
	if strings.Index(w.Body.String(), "a-bucket") > strings.Index(w.Body.String(), "z-bucket") {
		t.Fatalf("not sorted: %s", w.Body.String())
	}
}
func TestZeroByteAndRange(t *testing.T) {
	h := setup(t)
	call(t, h, "PUT", "/bucket-one", "CreateBucket", "")
	call(t, h, "PUT", "/bucket-one/empty", "PutObject", "")
	w := call(t, h, "GET", "/bucket-one/empty", "GetObject", "")
	if w.Code != 200 || w.Header().Get("Content-Length") != "0" {
		t.Fatalf("zero: %d %v", w.Code, w.Header())
	}
	r := httptest.NewRequest("GET", "/bucket-one/empty", nil)
	r.Header.Set("Range", "bytes=0-1")
	req := &awsprovider.Request{HTTPRequest: r, Protocol: awsprovider.ProtocolS3, Operation: "GetObject"}
	w = httptest.NewRecorder()
	h.ServeAWS(w, req, "id")
	if w.Code != 501 {
		t.Fatalf("range=%d", w.Code)
	}
}
