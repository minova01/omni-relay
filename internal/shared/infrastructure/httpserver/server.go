package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"omni-relay/configs"
)

type Server struct {
	http *http.Server
}

func New(cfg configs.HTTPConfig, logger *slog.Logger, dependencies []Dependency) *Server {
	health := newHealthHandler(dependencies)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", health.live)
	mux.HandleFunc("GET /health/ready", health.ready)

	return &Server{http: &http.Server{
		Addr:         cfg.Address,
		Handler:      middleware(logger, mux),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}}
}

func (s *Server) Run() error {
	err := s.http.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
