// internal/indexer/processor/metrics.go
package processor

/*@
 * metrics.go
 * @desc Placeholder metrics for retry processor
 * @note Replace with actual Prometheus metrics when metrics package is ready
 */

// ProcessingErrors is a placeholder for error metrics
var ProcessingErrors = &placeholderCounterVec{}

// EventsProcessed is a placeholder for event metrics
var EventsProcessed = &placeholderCounterVec{}

// placeholderCounter implements a no-op counter
type placeholderCounter struct{}

func (p *placeholderCounter) Inc() {}

func (p *placeholderCounter) Add(float64) {}

// placeholderCounterVec implements a no-op counter vector
type placeholderCounterVec struct{}

func (p *placeholderCounterVec) WithLabelValues(labels ...string) *placeholderCounter {
    return &placeholderCounter{}
}