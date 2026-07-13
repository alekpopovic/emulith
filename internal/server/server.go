package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/alekpopovic/emulith/internal/service"
	"github.com/alekpopovic/emulith/internal/state"
)

type Server struct {
	httpServer *http.Server
}

func New(addr, version string, rootHandlers ...http.Handler) *Server {
	return newServer(addr, version, nil, rootHandlers...)
}
func newServer(addr, version string, registry *service.Registry, rootHandlers ...http.Handler) *Server {
	mux := http.NewServeMux()
	mux.Handle("GET /_emulith/health", HealthHandler(version, registry))
	if len(rootHandlers) > 0 && rootHandlers[0] != nil {
		mux.Handle("/", rootHandlers[0])
	}
	return &Server{httpServer: &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}}
}

func NewWithState(addr, version string, store *state.Store, logger *slog.Logger, root http.Handler, registries ...*service.Registry) *Server {
	var registry *service.Registry
	if len(registries) > 0 {
		registry = registries[0]
	}
	s := newServer(addr, version, registry, root)
	s.httpServer.Handler.(*http.ServeMux).Handle("POST /_emulith/reset", ResetHandler(store, logger, registry))
	s.httpServer.Handler.(*http.ServeMux).HandleFunc("/_emulith/reset", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	s.httpServer.Handler.(*http.ServeMux).HandleFunc("GET /_emulith/state/export", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/gzip")
		w.Header().Set("Content-Disposition", `attachment; filename="emulith-state.tar.gz"`)
		if err := store.Export(r.Context(), w, version, time.Now()); err != nil {
			logger.Error("state export failed", "error", err)
		}
	})
	s.httpServer.Handler.(*http.ServeMux).HandleFunc("POST /_emulith/state/import", func(w http.ResponseWriter, r *http.Request) {
		if err := store.Import(r.Context(), http.MaxBytesReader(w, r.Body, 4<<30), r.URL.Query().Get("replace") == "true"); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			logger.Warn("state import rejected", "error", err)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "snapshot import rejected"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "imported": true})
	})
	return s
}

func (s *Server) ListenAndServe() error    { return s.httpServer.ListenAndServe() }
func (s *Server) HTTPServer() *http.Server { return s.httpServer }

func HealthHandler(version string, registries ...*service.Registry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		response := struct {
			Status   string                          `json:"status"`
			Name     string                          `json:"name"`
			Version  string                          `json:"version"`
			Services map[string]service.HealthStatus `json:"services,omitempty"`
		}{Status: "ok", Name: "emulith", Version: version}
		if len(registries) > 0 && registries[0] != nil {
			response.Services = registries[0].Health(r.Context())
			for _, h := range response.Services {
				if h.Status != "ok" {
					response.Status = "degraded"
					w.WriteHeader(http.StatusServiceUnavailable)
					break
				}
			}
		}
		_ = json.NewEncoder(w).Encode(response)
	})
}

func ResetHandler(store *state.Store, logger *slog.Logger, registries ...*service.Registry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		requestID := r.Header.Get("x-amzn-RequestId")
		if requestID == "" {
			requestID = "admin-local"
		}
		result := "ok"
		defer func() {
			logger.Info("admin reset", "request_id", requestID, "result", result, "duration_ms", time.Since(started).Milliseconds())
		}()
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if len(registries) > 0 && registries[0] != nil {
			if err := registries[0].Reset(ctx); err != nil {
				result = "error"
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(500)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": "error", "reset": false})
				return
			}
		}
		if err := store.Reset(ctx); err != nil {
			result = "error"
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(500)
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "error", "reset": false})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "reset": true})
	})
}
