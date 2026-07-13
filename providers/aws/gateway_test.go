package aws

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alekpopovic/emulith/internal/server"
	"github.com/alekpopovic/emulith/internal/state"
)

func testGateway(t *testing.T) (*Gateway, *bytes.Buffer) {
	t.Helper()
	store, err := state.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	var logs bytes.Buffer
	return NewGateway(store, slog.New(slog.NewJSONHandler(&logs, nil))), &logs
}

func TestClassification(t *testing.T) {
	tests := []struct{ name, method, target, contentType, body, service, operation string }{
		{"sqs json", "POST", "/", "application/x-amz-json-1.0; charset=utf-8", "{}", "sqs", "CreateQueue"},
		{"sqs query", "POST", "/", "application/x-www-form-urlencoded", "Action=CreateQueue&Version=2012-11-05", "sqs", "CreateQueue"},
		{"sts query", "GET", "/?Action=GetCallerIdentity&Version=2011-06-15", "", "", "sts", "GetCallerIdentity"},
		{"bucket", "PUT", "/my-bucket", "", "", "s3", "CreateBucket"},
		{"object", "PUT", "/my-bucket/a.txt", "", "", "s3", "PutObject"},
		{"list", "GET", "/my-bucket?list-type=2", "", "", "s3", "ListObjectsV2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.target, strings.NewReader(tt.body))
			if tt.contentType != "" {
				r.Header.Set("Content-Type", tt.contentType)
			}
			if tt.name == "sqs json" {
				r.Header.Set("X-Amz-Target", "AmazonSQS.CreateQueue")
			}
			got, err := classify(r)
			if err != nil {
				t.Fatal(err)
			}
			if got.Service != tt.service || got.Operation != tt.operation {
				t.Fatalf("got %s/%s", got.Service, got.Operation)
			}
		})
	}
}

func TestResponseShapesAndSafeLogging(t *testing.T) {
	gateway, logs := testGateway(t)
	tests := []struct{ name, method, target, contentType, amzTarget, wantType, wantHeader string }{
		{"json", "POST", "/", "application/x-amz-json-1.0", "AmazonSQS.Unknown", "application/x-amz-json-1.0", "x-amzn-RequestId"},
		{"query", "POST", "/", "application/x-www-form-urlencoded", "", "text/xml", "x-amzn-RequestId"},
		{"s3", "PUT", "/bucket", "", "", "application/xml", "x-amz-request-id"},
	}
	for _, tt := range tests {
		r := httptest.NewRequest(tt.method, tt.target, strings.NewReader("Action=Unknown&MessageBody=secret"))
		r.Header.Set("Content-Type", tt.contentType)
		r.Header.Set("X-Amz-Target", tt.amzTarget)
		r.Header.Set("Authorization", "SECRET-AUTH")
		w := httptest.NewRecorder()
		gateway.ServeHTTP(w, r)
		if !strings.HasPrefix(w.Header().Get("Content-Type"), tt.wantType) || w.Header().Get(tt.wantHeader) == "" {
			t.Fatalf("headers=%v body=%s", w.Header(), w.Body.String())
		}
	}
	if strings.Contains(logs.String(), "SECRET-AUTH") || strings.Contains(logs.String(), "MessageBody") || strings.Contains(logs.String(), "secret") {
		t.Fatalf("sensitive log: %s", logs.String())
	}
}

func TestAdminWinsAndOversizeRejected(t *testing.T) {
	gateway, _ := testGateway(t)
	app := server.New(":0", "dev", gateway).HTTPServer().Handler
	w := httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/_emulith/health", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"status":"ok"`) {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	body := io.LimitReader(strings.NewReader(strings.Repeat("x", maxProtocolBody+1)), maxProtocolBody+1)
	r := httptest.NewRequest(http.MethodPost, "/", body)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	gateway.ServeHTTP(w, r)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
