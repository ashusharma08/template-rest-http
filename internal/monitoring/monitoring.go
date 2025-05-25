package monitoring

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

var (
	monitor *Monitor
	once    sync.Once
)

type Monitor struct {
	ready bool
	live  bool
	lock  sync.Mutex
}

func GetMonitor() *Monitor {
	once.Do(func() {
		monitor = &Monitor{
			ready: false,
			live:  true,
		}
	})
	return monitor
}

func (m *Monitor) SetReady(ready bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.ready = ready
}

func (m *Monitor) SetLive(live bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.live = live
}

func (m *Monitor) GetReady() bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.ready
}

func (m *Monitor) GetLive() bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.live
}

func (m *Monitor) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	srvMonitor := &http.Server{
		Addr:    ":9090",
		Handler: m,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	go func() {
		slog.Info("Starting monitor server on port :9090")
		if err := srvMonitor.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				// cannot panic, because this probably is an intentional close
				slog.Info("Monitor server stopped")
			} else {
				slog.Error("error starting monitoring server ", "error", err)
			}
		}
	}()
	// wait for shutdown
	<-ctx.Done()
	timeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := srvMonitor.Shutdown(timeout)
	if err != nil {
		slog.Error("Error shutting down monitoring server", "error", err)
	}
}

func (m *Monitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/metrics":
		//handle metrics
	case "/ready":
		if m.ready {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
		}
	case "/live":
		if m.live {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
		}
	}
}
