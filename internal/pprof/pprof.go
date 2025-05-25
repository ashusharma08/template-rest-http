package pprof

import (
	"context"
	"log/slog"
	"net/http"
	netpprof "net/http/pprof"
	"sync"
	"time"
)

type Pprof struct {
}

func New() *Pprof {
	return &Pprof{}
}

func (p *Pprof) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	srvPprof := &http.Server{
		Addr:    ":7777",
		Handler: p,
	}
	go func() {
		slog.Info("Starting pprof server on port :7777")
		if err := srvPprof.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				// cannot panic, because this probably is an intentional close
				slog.Info("Pprof server stopped")
			} else {
				slog.Error("Pprof server error occurred", "error", err)
				return
			}
		}
	}()
	// wait for shutdown
	<-ctx.Done()
	timeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := srvPprof.Shutdown(timeout)
	if err != nil {
		slog.Error("Error shutting down pprof server", "error", err)
	}
}

func (p *Pprof) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/debug/pprof":
		http.HandlerFunc(netpprof.Index).ServeHTTP(w, r)
	case "/debug/cmdline":
		http.HandlerFunc(netpprof.Cmdline).ServeHTTP(w, r)
	case "/debug/symbol":
		http.HandlerFunc(netpprof.Symbol).ServeHTTP(w, r)
	case "/debug/heap":
		http.HandlerFunc(netpprof.Handler("heap").ServeHTTP).ServeHTTP(w, r)
	case "/debug/goroutine":
		http.HandlerFunc(netpprof.Handler("goroutine").ServeHTTP).ServeHTTP(w, r)
	case "/debug/profile":
		http.HandlerFunc(netpprof.Profile).ServeHTTP(w, r)
	case "/debug/block":
		http.HandlerFunc(netpprof.Handler("block").ServeHTTP).ServeHTTP(w, r)
	case "/debug/threadcreate":
		http.HandlerFunc(netpprof.Handler("threadcreate").ServeHTTP).ServeHTTP(w, r)
	case "/debug/trace":
		http.HandlerFunc(netpprof.Trace).ServeHTTP(w, r)
	default:
		http.HandlerFunc(netpprof.Index).ServeHTTP(w, r)

	}
}
