package metrics

import (
	"context"
	"fmt"
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

	mempoolTransactions = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "monero_mempool_transactions",
		Help: "number of transactions in the mempool",
	})
)

type Collector struct {
	client   *daemonrpc.Client
	interval time.Duration
	log      logr.Logger
}

const CollectionInterval = 10 * time.Second

func NewCollector(client *daemonrpc.Client) *Collector {
	return &Collector{
		client:   client,
		interval: CollectionInterval,
		log:      log.Log.WithName("collector"),
	}
}

func (c *Collector) Run(ctx context.Context) error {
	ticker := time.NewTicker(c.interval)

	if err := c.Collect(ctx); err != nil {
		return fmt.Errorf("collect: %w", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := c.Collect(ctx); err != nil {
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

func (c *Collector) Collect(ctx context.Context) error {
	c.log.V(7).Info("collection-start")
	defer c.log.V(7).Info("collection-finished")

	funcs := []CollectorFunc{
		c.CollectBlockCount,
	}

	for idx, f := range funcs {
		if err := f(ctx); err != nil {
			return fmt.Errorf("collect f %d: %w", idx, err)
		}
	}

	return nil
}

func (c *Collector) CollectBlockCount(ctx context.Context) error {
	c.log.V(7).Info("collect-block-count")

	res, err := c.client.GetBlockCount(ctx)
	if err != nil {
		return fmt.Errorf("get block count: %w", err)
	}

	blockCount.Set(float64(res.Count))
	return nil
}

func (c *Collector) CollectFeeEstimate(ctx context.Context) error {
	c.log.V(7).Info("collect-fee-estimate")

	res, err := c.client.GetFeeEstimate(ctx, uint64(1))
	if err != nil {
		return fmt.Errorf("get fee estimate: %w", err)
	}

	feeEstimate.Set(float64(res.Fee))
	return nil
}

func (c *Collector) CollectMempoolTransactions(ctx context.Context) error {
	c.log.V(7).Info("collect-mempool-transactions")

	res, err := c.client.GetTransactionPoolStats(ctx)
	if err != nil {
		return fmt.Errorf("get transaction pool: %w", err)
	}

	mempoolTransactions.Set(float64(res.PoolStats.TxsTotal))
	return nil
}
