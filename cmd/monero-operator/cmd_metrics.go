package main

import (
	"context"
	"fmt"
	"net"

	"github.com/cirocosta/go-monero/pkg/daemonrpc"
	"github.com/oschwald/geoip2-golang"

	"github.com/cirocosta/monero-operator/pkg/metrics"
)

type MetricsCommand struct {
	MonerodAddress string `long:"monerod-address" required:"true"`
	GeoIPFile      string `long:"geoip-file" default:"./hack/geoip.mmdb"`
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

	db, err := geoip2.Open(c.GeoIPFile)
	if err != nil {
		return fmt.Errorf("geoip open: %w", err)
	}
	defer db.Close()

	countryMapper := func(ip net.IP) (string, error) {
		res, err := db.Country(ip)
		if err != nil {
			return "", fmt.Errorf("country '%s': %w", ip, err)
		}

		return res.RegisteredCountry.IsoCode, nil
	}

	if err := metrics.RegisterCollector(daemonClient, metrics.WithCountryMapper(countryMapper)); err != nil {
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
