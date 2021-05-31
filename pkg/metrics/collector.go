package metrics

import (
	"context"
	"fmt"
	"math"
	"net"
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

type CountryMapper func(net.IP) (string, error)

type Collector struct {
	client *daemonrpc.Client
	log    logr.Logger

	countryMapper CountryMapper
}

type CollectorOpt func(c *Collector)

func WithCountryMapper(v CountryMapper) func(c *Collector) {
	return func(c *Collector) {
		c.countryMapper = v
	}
}

func RegisterCollector(client *daemonrpc.Client, opts ...CollectorOpt) error {
	c := &Collector{
		client:        client,
		log:           log.Log.WithName("collector"),
		countryMapper: func(_ net.IP) (string, error) { return "lol", nil },
	}

	for _, opt := range opts {
		opt(c)
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
		{"fee_estimate", c.CollectFeeEstimate},
		{"peers", c.CollectPeers},
		{"connections", c.CollectConnections},
		{"last_block_stats", c.CollectLastBlockStats},
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

func (c *Collector) CollectConnections(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetConnections(ctx)
	if err != nil {
		return fmt.Errorf("get connections: %w", err)
	}

	perCountryCounter := map[string]uint64{}
	for _, conn := range res.Connections {
		country, err := c.countryMapper(net.ParseIP(conn.Host))
		if err != nil {
			return fmt.Errorf("to country '%s': %w", conn.Host, err)
		}

		perCountryCounter[country] += 1
	}

	desc := prometheus.NewDesc(
		"monero_connections",
		"connections info",
		[]string{"country"}, nil,
	)

	for country, count := range perCountryCounter {
		ch <- prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			float64(count),
			country,
		)
	}

	return nil
}

func (c *Collector) CollectPeers(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetPeerList(ctx)
	if err != nil {
		return fmt.Errorf("get peer list: %w", err)
	}

	perCountryCounter := map[string]uint64{}
	for _, peer := range res.WhiteList {
		country, err := c.countryMapper(net.ParseIP(peer.Host))
		if err != nil {
			return fmt.Errorf("to country '%s': %w", peer.Host, err)
		}

		perCountryCounter[country] += 1
	}

	desc := prometheus.NewDesc(
		"monero_peers_new",
		"peers info",
		[]string{"country"}, nil,
	)

	for country, count := range perCountryCounter {
		ch <- prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			float64(count),
			country,
		)
	}

	return nil
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

func (c *Collector) CollectLastBlockStats(ctx context.Context, ch chan<- prometheus.Metric) error {
	lastBlockHeaderResp, err := c.client.GetLastBlockHeader(ctx)
	if err != nil {
		return fmt.Errorf("get last block header: %w", err)
	}

	currentHeight := lastBlockHeaderResp.BlockHeader.Height

	block, err := c.client.GetBlock(ctx, currentHeight)
	if err != nil {
		return fmt.Errorf("get block '%d': %w", currentHeight, err)
	}

	blockJson, err := block.InnerJSON()
	if err != nil {
		return fmt.Errorf("block inner json: %w", err)
	}

	txnsResp, err := c.client.GetTransactions(ctx, blockJson.TxHashes)
	if err != nil {
		return fmt.Errorf("get txns: %w", err)
	}

	txns, err := txnsResp.GetTransactions()
	if err != nil {
		return fmt.Errorf("get transactions: %w", err)
	}

	phis := []float64{0.25, 0.50, 0.75, 0.90, 0.95, 0.99, 1}

	var (
		streamVin    = quantile.NewTargeted(phis...)
		sumVin       = float64(0)
		quantilesVin = make(map[float64]float64, len(phis))

		streamVout    = quantile.NewTargeted(phis...)
		sumVout       = float64(0)
		quantilesVout = make(map[float64]float64, len(phis))
	)

	for _, txn := range txns {
		streamVin.Insert(float64(len(txn.Vin)))
		sumVin += float64(len(txn.Vin))

		streamVout.Insert(float64(len(txn.Vout)))
		sumVout += float64(len(txn.Vout))
	}

	for _, phi := range phis {
		quantilesVin[phi] = streamVin.Query(phi)
		quantilesVout[phi] = streamVout.Query(phi)
	}

	ch <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_last_block_vout",
			"distribution of outputs in last block",
			nil, nil,
		),
		uint64(streamVout.Count()),
		sumVout,
		quantilesVout,
	)

	ch <- prometheus.MustNewConstSummary(
		prometheus.NewDesc(
			"monero_last_block_vin",
			"distribution of inputs in last block",
			nil, nil,
		),
		uint64(streamVin.Count()),
		sumVin,
		quantilesVin,
	)

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

func (c *Collector) CollectFeeEstimate(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetFeeEstimate(ctx, 1)
	if err != nil {
		return fmt.Errorf("get fee estimate: %w", err)
	}

	desc := prometheus.NewDesc(
		"monero_fee_estimate",
		"fee estimate for 1 grace block",
		nil, nil,
	)

	ch <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(res.Fee),
	)

	return nil
}

func (c *Collector) CollectBans(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetBans(ctx)
	if err != nil {
		return fmt.Errorf("get bans: %w", err)
	}

	desc := prometheus.NewDesc(
		"monero_bans",
		"number of nodes banned",
		nil, nil,
	)

	ch <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		float64(len(res.Bans)),
	)

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
