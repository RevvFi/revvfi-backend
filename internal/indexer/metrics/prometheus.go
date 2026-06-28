// internal/indexer/metrics/prometheus.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    EventsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "revvfi_indexer_events_processed_total",
        Help: "Total number of events processed by type",
    }, []string{"event_name"})

    BlocksProcessed = promauto.NewCounter(prometheus.CounterOpts{
        Name: "revvfi_indexer_blocks_processed_total",
        Help: "Total number of blocks processed",
    })

    CurrentBlockLag = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "revvfi_indexer_block_lag",
        Help: "Number of blocks behind chain head",
    })

    ReorgsDetected = promauto.NewCounter(prometheus.CounterOpts{
        Name: "revvfi_indexer_reorgs_detected_total",
        Help: "Total number of chain reorganizations detected",
    })

    ProcessingErrors = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "revvfi_indexer_processing_errors_total",
        Help: "Total number of processing errors by type",
    }, []string{"error_type"})
)

func InitMetrics() {
    // Register default metrics (already done by promauto)
}
