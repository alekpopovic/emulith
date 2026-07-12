package server

import (
	"encoding/json"
	"net/http"
	"time"
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

func (s *Server) ListenAndServe() error    { return s.httpServer.ListenAndServe() }
func (s *Server) HTTPServer() *http.Server { return s.httpServer }

func HealthHandler(version string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "name": "emulith", "version": version})
	})
}
