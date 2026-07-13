package sts

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	awsprovider "github.com/alekpopovic/emulith/providers/aws"
)

func TestGetCallerIdentity(t *testing.T) {
	form := url.Values{"Action": {"GetCallerIdentity"}, "Version": {"2011-06-15"}}
	req := &awsprovider.Request{HTTPRequest: httptest.NewRequest(http.MethodPost, "/", nil), Protocol: awsprovider.ProtocolQuery, Service: "sts", Operation: "GetCallerIdentity", Form: form}
	w := httptest.NewRecorder()
	New().ServeAWS(w, req, "request-1")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "EMULITHUSER") || !strings.Contains(w.Body.String(), namespace) || w.Header().Get("x-amzn-RequestId") != "request-1" {
		t.Fatalf("status=%d headers=%v body=%s", w.Code, w.Header(), w.Body.String())
	}
}

func TestInvalidAction(t *testing.T) {
	req := &awsprovider.Request{Protocol: awsprovider.ProtocolQuery, Operation: "AssumeRole", Form: url.Values{}}
	w := httptest.NewRecorder()
	New().ServeAWS(w, req, "request-2")
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "InvalidAction") {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
