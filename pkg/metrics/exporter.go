package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Exporter struct {
	// ListenAddress is the full address used by prometheus
	// to listen for scraping requests.
	//
	// Examples:
	// - :8080
	// - 127.0.0.2:1313
	//
	listenAddress string

	// TelemetryPath configures the path under which
	// the prometheus metrics are reported.
	//
	// For instance:
	// - /metrics
	// - /telemetry
	//
	telemetryPath string

	listener net.Listener
	log      logr.Logger
}

type Option func(e *Exporter)

func NewExporter(opts ...Option) *Exporter {
	e := &Exporter{
		listenAddress: ":9000",
		telemetryPath: "/metrics",
		log:           log.Log.WithName("exporter"),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Run initiates the HTTP server to serve the metrics.
//
// This is a blocking method - make sure you either make use of goroutines to
// not block if needed.
//
func (e *Exporter) Run(ctx context.Context) error {

	var err error

	e.listener, err = net.Listen("tcp", e.listenAddress)
	if err != nil {
		return fmt.Errorf("listen on '%s': %w", e.listenAddress, err)
	}

	doneChan := make(chan error, 1)

	go func() {
		defer close(doneChan)

		http.Handle(e.telemetryPath, promhttp.Handler())
		if err := http.Serve(e.listener, nil); err != nil {
			doneChan <- fmt.Errorf(
				"failed listening on address %s: %w",
				e.listenAddress, err,
			)
		}
	}()

	select {
	case err = <-doneChan:
		if err != nil {
			return fmt.Errorf("donechan err: %w", err)
		}
	case <-ctx.Done():
		return fmt.Errorf("ctx err: %w", ctx.Err())
	}

	return nil
}

// Close gracefully closes the tcp listener associated with it.
//
func (e *Exporter) Close() (err error) {
	if e.listener == nil {
		return nil
	}

	if err := e.listener.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}

	return nil
}
