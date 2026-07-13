package dynamodb

import (
	awsprovider "github.com/alekpopovic/emulith/providers/aws"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUnsupportedEnvelope(t *testing.T) {
	w := httptest.NewRecorder()
	New(nil).ServeAWS(w, &awsprovider.Request{Operation: "Unknown"}, "id")
	if w.Code != 400 || w.Header().Get("x-amzn-RequestId") != "id" || !strings.Contains(w.Body.String(), "UnknownOperationException") {
		t.Fatal(w.Code, w.Body.String())
	}
}
