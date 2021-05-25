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

	if err := metrics.RegisterCollector(daemonClient); err != nil {
		return fmt.Errorf("new collector: %w", err)
	}

	if err := exporter.Run(ctx); err != nil {
		return fmt.Errorf("exporter run: %w", err)
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
