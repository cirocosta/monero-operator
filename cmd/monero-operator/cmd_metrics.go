package main

import (
	"context"
	"fmt"

	"github.com/cirocosta/go-monero/pkg/daemonrpc"
	"github.com/cirocosta/monero-operator/pkg/metrics"
)

type MetricsCommand struct {
	MonerodAddress string `long:"monerod-address" required:"true"`
}

func (c *MetricsCommand) Execute(_ []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exporter := metrics.NewExporter()
	defer exporter.Close()

	daemonClient, err := daemonrpc.NewClient(c.MonerodAddress)
	if err != nil {
		return fmt.Errorf("new client '%s': %w", c.MonerodAddress, err)
	}

	collector := metrics.NewCollector(daemonClient)

	exporterC := make(chan error)
	go func() {
		defer close(exporterC)

		if err := exporter.Run(ctx); err != nil {
			exporterC <- fmt.Errorf("listen: %w", err)
		}
	}()

	collectorC := make(chan error)
	go func() {
		defer close(collectorC)

		if err := collector.Run(ctx); err != nil {
			collectorC <- fmt.Errorf("run: %w", err)
		}
	}()

	select {
	case err := <-exporterC:
		if err != nil {
			return fmt.Errorf("exporterC: %w", err)
		}
	case err := <-collectorC:
		if err != nil {
			return fmt.Errorf("collectorC: %w", err)
		}
	}

	return nil
}

func init() {
	parser.AddCommand("metrics",
		"grab metrics from monerod",
		"grab metrics from monerod",
		&MetricsCommand{},
	)
}
