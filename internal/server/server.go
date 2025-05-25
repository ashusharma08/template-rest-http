package server

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/esoptra/rest-http/config"
	"github.com/esoptra/rest-http/internal/monitoring"
	"github.com/esoptra/rest-http/internal/server/router"
)

type Server struct{}

func NewServer(cfg *config.Config) *Server {
	return &Server{}
}

func (s *Server) Start(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	defer wg.Done()

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router.NewRouter(),  
		WriteTimeout: 600 * time.Second,
		ReadTimeout:  600 * time.Second,
	}

	go func() {
		slog.Info("Starting server on port :80")
		if err := srv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed { 
				slog.Info("application server stopped", "error", err)
			} else {
				slog.Error("application server error occurred", "error", err)
				return
			}
		}
	}()
	monitoring.GetMonitor().SetReady(true)
	<-ctx.Done()
	monitoring.GetMonitor().SetReady(false)
	timeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := srv.Shutdown(timeout)
	if err != nil {
		slog.Error("Error shutting down application server", "error", err)
	}
	return nil
}
