package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/emulith/emulith/internal/state"
)

type Server struct {
	httpServer *http.Server
}

func New(addr, version string, rootHandlers ...http.Handler) *Server {
	mux := http.NewServeMux()
	mux.Handle("GET /_emulith/health", HealthHandler(version))
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

func NewWithState(addr, version string, store *state.Store, logger *slog.Logger, root http.Handler) *Server {
	s := New(addr, version, root)
	s.httpServer.Handler.(*http.ServeMux).Handle("POST /_emulith/reset", ResetHandler(store, logger))
	s.httpServer.Handler.(*http.ServeMux).HandleFunc("/_emulith/reset", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	return s
}

func (s *Server) ListenAndServe() error    { return s.httpServer.ListenAndServe() }
func (s *Server) HTTPServer() *http.Server { return s.httpServer }

func HealthHandler(version string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "name": "emulith", "version": version})
	})
}

func ResetHandler(store *state.Store, logger *slog.Logger) http.Handler {
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
