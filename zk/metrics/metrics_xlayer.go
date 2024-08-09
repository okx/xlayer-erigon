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
)

func XLayerMetricsInit() {
	prometheus.MustRegister(BatchExecuteTimeGauge)
}

var BatchExecuteTimeGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: BatchExecuteTimeName,
		Help: "[SEQUENCER] batch execution time in second",
	},
	[]string{"closingReason"},
)

// BatchExecuteTime sets the gauge vector to the given batch type and time.
func BatchExecuteTime(closingReason string, duration time.Duration) {
	log.Info(fmt.Sprintf("[BatchExecuteTime] ClosingReason: %s, Duration: %.2fs", closingReason, duration.Seconds()))
	BatchExecuteTimeGauge.WithLabelValues(closingReason).Set(duration.Seconds())
}
