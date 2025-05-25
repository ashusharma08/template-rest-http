package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/esoptra/rest-http/config"
	"github.com/esoptra/rest-http/internal/monitoring"
	"github.com/esoptra/rest-http/internal/pprof"
	"github.com/esoptra/rest-http/internal/server"
)

func initLogger() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func main() {
	initLogger()
	if err := mainErr(); err != nil {
		slog.Error("error starting server ", "error", err)
	}
}

func mainErr() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	//start monitoring
	m := monitoring.GetMonitor()
	go m.Start(ctx, &wg)
	//start pprof
	go pprof.New().Start(ctx, &wg)
	//start server
	s := server.NewServer(cfg)
	go s.Start(ctx, &wg)

	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT)
	signal.Notify(signals, syscall.SIGTERM)

	<-signals
	cancel()
	wg.Wait()
	slog.Info("shutdown complete")
	return nil
}
