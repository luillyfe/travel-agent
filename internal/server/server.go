package server

import (
	"net/http"
	"travel-agent/internal/config"
)

type Server struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Server {
	return &Server{
		cfg: cfg,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/health", s.HealthHandler)
	return http.ListenAndServe(s.cfg.ServerPort, nil)
}

// HealthHandler responds with a JSON indicating the server is healthy.
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"OK"}`)); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
