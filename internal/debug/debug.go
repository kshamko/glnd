// Package debug provide mechanics for serve debug information such as metrics,pprof,healthz ...
package debug

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	health "github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/handlers"
	"github.com/kshamko/glnd/internal/restapi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Service for debugging.
type Service struct {
	healthd *health.Health
}

// New debug interface.
func New(healthd *health.Health) *Service {
	return &Service{
		healthd: healthd,
	}
}

// Serve debug information. It doesn't use standart http server.
func (s *Service) Serve(ctx context.Context, l string) error {
	if err := s.healthd.Start(); err != nil {
		return err
	}

	r := http.NewServeMux()
	r.Handle("/metrics", promhttp.Handler())
	r.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	r.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	r.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	r.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	r.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	r.HandleFunc("/healthz", handlers.NewJSONHandlerFunc(s.healthd, nil))

	fs := http.FileServer(http.Dir("assets/swagger-ui"))
	r.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", fs))
	r.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, string(restapi.FlatSwaggerJSON))
	})

	ms := http.Server{
		Addr:         l,
		Handler:      r,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	}
	e := make(chan error, 1)

	go func() {
		e <- ms.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return ms.Shutdown(context.Background()) //nolint: contextcheck
	case err := <-e:
		return err
	}
}
