package metrics

import (
	"fmt"
	"github.com/ledgerwatch/log/v3"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

// BatchFinalizeTypeLabel batch finalize type label
type BatchFinalizeTypeLabel string

const (
	// BatchFinalizeTypeLabelDeadline batch finalize type deadline label
	BatchFinalizeTypeLabelDeadline BatchFinalizeTypeLabel = "deadline"
	// BatchFinalizeTypeLabelFullBatch batch finalize type full batch label
	BatchFinalizeTypeLabelFullBatch BatchFinalizeTypeLabel = "full_batch"
)

const (
	CounterOverflow = "BatchCounterOverflow"
	EmptyTimeOut    = "EmptyBatchTimeOut"
	NonEmptyTimeOut = "NonEmptyBatchTimeOut"
)

var (
	SeqPrefix = "sequencer_"
	// BatchExecuteTimeName is the name of the metric that shows the batch execution time.
	BatchExecuteTimeName = SeqPrefix + "batch_execute_time"
	PoolTxCountName      = SeqPrefix + "pool_tx_count"
)

func XLayerMetricsInit() {
	prometheus.MustRegister(BatchExecuteTimeGauge)
	prometheus.MustRegister(PoolTxCount)
}

var BatchExecuteTimeGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: BatchExecuteTimeName,
		Help: "[SEQUENCER] batch execution time in second",
	},
	[]string{"closingReason"},
)

var PoolTxCount = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: PoolTxCountName,
		Help: "[SEQUENCER] txcount of each pool in txpool",
	},
	[]string{"poolName"},
)

// BatchExecuteTime sets the gauge vector to the given batch type and time.
func BatchExecuteTime(closingReason string, duration time.Duration) {
	log.Info(fmt.Sprintf("[BatchExecuteTime] ClosingReason: %s, Duration: %.2fs", closingReason, duration.Seconds()))
	BatchExecuteTimeGauge.WithLabelValues(closingReason).Set(duration.Seconds())
}

func AddPoolTxCount(pending, basefee, queued int) {
	log.Info(fmt.Sprintf("[PoolTxCount] pending: %d, basefee: %d, queued: %d", pending, basefee, queued))
	PoolTxCount.WithLabelValues("pending").Set(float64(pending))
	PoolTxCount.WithLabelValues("basefee").Set(float64(basefee))
	PoolTxCount.WithLabelValues("queued").Set(float64(queued))
}
