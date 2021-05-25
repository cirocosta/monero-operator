package metrics

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"

	"github.com/bmizerany/perks/quantile"
	"github.com/cirocosta/go-monero/pkg/daemonrpc"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Collector struct {
	client *daemonrpc.Client
	log    logr.Logger
}

func RegisterCollector(client *daemonrpc.Client) error {
	c := &Collector{
		client: client,
		log:    log.Log.WithName("collector"),
	}

	if err := prometheus.Register(c); err != nil {
		return fmt.Errorf("register: %w", err)
	}

	return nil
}

type CollectorFunc func(ctx context.Context, ch chan<- prometheus.Metric) error

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	return
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	var g *errgroup.Group

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	g, ctx = errgroup.WithContext(ctx)

	for _, collector := range []struct {
		name string
		fn   CollectorFunc
	}{
		{"info_stats", c.CollectInfoStats},
		{"mempool_stats", c.CollectMempoolStats},
		{"last_block_header", c.CollectLastBlockHeader},
		{"bans", c.CollectBans},
		{"peer_height_divergence", c.CollectPeerHeightDivergence},
	} {
		collector := collector

		g.Go(func() error {
			if err := collector.fn(ctx, ch); err != nil {
				return fmt.Errorf("collector fn '%s': %w", collector.name, err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		c.log.Error(err, "wait")
	}

	return
}

func (c *Collector) CollectLastBlockHeader(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetLastBlockHeader(ctx)
	if err != nil {
		return fmt.Errorf("get last block header: %w", err)
	}

	metrics, err := c.toMetrics("last_block_header", &res.BlockHeader)
	if err != nil {
		return fmt.Errorf("to metrics: %w", err)
	}

	for _, metric := range metrics {
		ch <- metric
	}

	return nil
}

func (c *Collector) CollectInfoStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetInfo(ctx)
	if err != nil {
		return fmt.Errorf("get transaction pool: %w", err)
	}

	metrics, err := c.toMetrics("info", res)
	if err != nil {
		return fmt.Errorf("to metrics: %w", err)
	}

	for _, metric := range metrics {
		ch <- metric
	}

	return nil
}

func (c *Collector) CollectPeerHeightDivergence(ctx context.Context, ch chan<- prometheus.Metric) error {
	blockCountRes, err := c.client.GetBlockCount(ctx)
	if err != nil {
		return fmt.Errorf("get block count: %w", err)
	}

	res, err := c.client.GetConnections(ctx)
	if err != nil {
		return fmt.Errorf("get connections: %w", err)
	}

	phis := []float64{0.25, 0.50, 0.55, 0.60, 0.65, 0.70, 0.75, 0.80, 0.85, 0.90, 0.95, 0.99}
	stream := quantile.NewTargeted(phis...)

	sum := float64(0)
	ourHeight := blockCountRes.Count
	for _, conn := range res.Connections {
		diff := math.Abs(float64(ourHeight - uint64(conn.Height)))

		stream.Insert(diff)
		sum += diff
	}

	quantiles := make(map[float64]float64, len(phis))
	for _, phi := range phis {
		quantiles[phi] = stream.Query(phi)
	}

	desc := prometheus.NewDesc(
		"monero_height_divergence",
		"how much our peers diverge from us in block height",
		nil, nil,
	)

	ch <- prometheus.MustNewConstSummary(
		desc,
		uint64(stream.Count()),
		sum,
		quantiles,
	)

	return nil
}

func (c *Collector) CollectBans(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetBans(ctx)
	if err != nil {
		return fmt.Errorf("get bans: %w", err)
	}

	var bans = struct {
		Count int `json:"count"`
	}{
		Count: len(res.Bans),
	}

	metrics, err := c.toMetrics("bans", &bans)
	if err != nil {
		return fmt.Errorf("to metrics: %w", err)
	}

	for _, metric := range metrics {
		ch <- metric
	}

	return nil
}

func (c *Collector) CollectMempoolStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetTransactionPoolStats(ctx)
	if err != nil {
		return fmt.Errorf("get transaction pool: %w", err)
	}

	metrics, err := c.toMetrics("mempool", &res.PoolStats)
	if err != nil {
		return fmt.Errorf("to metrics: %w", err)
	}

	for _, metric := range metrics {
		ch <- metric
	}

	return nil
}

func (c *Collector) toMetrics(ns string, res interface{}) ([]prometheus.Metric, error) {
	var (
		metrics = []prometheus.Metric{}
		v       = reflect.ValueOf(res).Elem()
		err     error
	)

	for i := 0; i < v.NumField(); i++ {
		observation := float64(0)

		field := v.Field(i)

		switch field.Type().Kind() {
		case
			reflect.Bool:
			if field.Bool() {
				observation = float64(1)
			}

		case
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.Uintptr:

			observation, err = strconv.ParseFloat(fmt.Sprintf("%v", field.Interface()), 64)
			if err != nil {
				return nil, fmt.Errorf("parse float: %w", err)
			}
		default:
			c.log.Info("ignoring",
				"field", v.Type().Field(i).Name,
				"type", field.Type().Kind().String(),
			)

			continue
		}

		tag := v.Type().Field(i).Tag.Get("json")

		metrics = append(metrics, prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				"monero_"+ns+"_"+tag,
				"info for "+tag,
				nil, nil,
			),
			prometheus.GaugeValue,
			observation,
		))
	}

	return metrics, nil
}
