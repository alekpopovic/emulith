package server

import (
	"context"
	"github.com/alekpopovic/emulith/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"
)

type healthService struct{ status string }

func (healthService) Provider() string { return "aws" }
func (healthService) Name() string     { return "fake" }
func (h healthService) Health(context.Context) service.HealthStatus {
	return service.HealthStatus{Status: h.status}
}
func (healthService) Reset(context.Context) error { return nil }
func TestAggregatedHealth(t *testing.T) {
	for _, tc := range []struct {
		status  string
		code    int
		overall string
	}{{"ok", 200, `"status":"ok"`}, {"error", 503, `"status":"degraded"`}} {
		registry := service.NewRegistry()
		registry.Register(healthService{tc.status})
		w := httptest.NewRecorder()
		HealthHandler("dev", registry).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/_emulith/health", nil))
		if w.Code != tc.code || !contains(w.Body.String(), tc.overall) || !contains(w.Body.String(), `"aws.fake"`) {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	}
}
func contains(value, part string) bool {
	return len(value) >= len(part) && func() bool {
		for i := 0; i+len(part) <= len(value); i++ {
			if value[i:i+len(part)] == part {
				return true
			}
		}
		return false
	}()
}
