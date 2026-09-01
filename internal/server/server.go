// Package server: HTTP server with graceful shutdown.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/stevensuki/ledgerline-backend/internal/config"
)

type Server struct {
	httpServer *http.Server
	log        *slog.Logger
	cfg        config.HTTP
}

func New(cfg config.HTTP, handler http.Handler, log *slog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              cfg.Addr(),
			Handler:           handler,
			ReadTimeout:       cfg.ReadTimeout,
			ReadHeaderTimeout: cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
		},
		log: log,
		cfg: cfg,
	}
}

// Run blocks until ctx is cancelled, then shuts down within a deadline.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.log.Info("http server listening", slog.String("addr", s.cfg.Addr()))
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("listen and serve: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.log.Info("shutdown signal received, closing the server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()

		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		s.log.Info("server stopped cleanly")
		return nil
	}
}
