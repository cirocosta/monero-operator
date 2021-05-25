package metrics

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/cirocosta/go-monero/pkg/daemonrpc"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	blockCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monero_block_count",
		Help: "number of blocks in the longest chain known to the node",
	})

	feeEstimate = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monero_fee_estimate",
		Help: "estimation on fee per byte in atomic units",
	})

	// TODO
	// mempoolTimeInPool = promauto.NewHistogram(prometheus.HistogramOpts{
	// 	Name:    "monero_mempool_time_in_pool_duration",
	// 	Help:    "todo",
	// 	Buckets: []float64{},
	// })
)

type Collector struct {
	client   *daemonrpc.Client
	interval time.Duration
	log      logr.Logger
}

const CollectionInterval = 10 * time.Second

func NewCollector(client *daemonrpc.Client) *Collector {
	c := &Collector{
		client:   client,
		interval: CollectionInterval,
		log:      log.Log.WithName("collector"),
	}

	prometheus.MustRegister(c)

	return c
}

func (c *Collector) Run(ctx context.Context) error {
	ticker := time.NewTicker(c.interval)

	if err := c.collect(ctx); err != nil {
		return fmt.Errorf("collect: %w", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := c.collect(ctx); err != nil {
				return fmt.Errorf("collect: %w", err)
			}
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("ctx err: %w", err)
			}
		}
	}

	return nil
}

type CollectorFunc func(ctx context.Context) error

func (c *Collector) collect(ctx context.Context) error {
	c.log.Info("collection-start")
	defer c.log.Info("collection-finished")

	for _, r := range []struct {
		name string
		f    CollectorFunc
	}{
		{
			"collect-block-count",
			c.CollectBlockCount,
		},
		// {
		// 	"collect-fee-estimate",
		// 	c.CollectFeeEstimate,	 -- TODO seems to HANG if we're not synced yet
		// },
		// {
		// 	"collect-mempool-transactions-count",
		// 	c.CollectMempoolStats,
		// },
	} {
		c.log.WithValues("collector", r.name).Info("collecting")
		if err := r.f(ctx); err != nil {
			return fmt.Errorf("collect f %s: %w", r.name, err)
		}
	}

	return nil
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	return
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	if err := c.CollectInfoStats(context.Background(), ch); err != nil {
		panic(err)
	}

	if err := c.CollectMempoolStats(context.Background(), ch); err != nil {
		panic(err)
	}

	if err := c.CollectLastBlockHeader(context.Background(), ch); err != nil {
		panic(err)
	}

	return
}

func (c *Collector) CollectBlockCount(ctx context.Context) error {
	res, err := c.client.GetBlockCount(ctx)
	if err != nil {
		return fmt.Errorf("get block count: %w", err)
	}

	blockCount.Set(float64(res.Count))
	return nil
}

func (c *Collector) CollectFeeEstimate(ctx context.Context) error {
	res, err := c.client.GetFeeEstimate(ctx, uint64(1))
	if err != nil {
		return fmt.Errorf("get fee estimate: %w", err)
	}

	feeEstimate.Set(float64(res.Fee))
	return nil
}

func (c *Collector) CollectLastBlockHeader(ctx context.Context, ch chan<- prometheus.Metric) error {
	res, err := c.client.GetLastBlockHeader(ctx)
	if err != nil {
		return fmt.Errorf("get last block header: %w", err)
	}

	metrics, err := c.toMetrics("last_block_header", res)
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
