// internal/indexer/processor/dispatcher.go
package processor

import (
    "context"
    "log"
    "sync"

    "github.com/ethereum/go-ethereum/core/types"

    "github.com/Revvfi/revvfi-backend/internal/indexer/decoder"
    "github.com/Revvfi/revvfi-backend/internal/handlers"
    "github.com/Revvfi/revvfi-backend/internal/indexer/idempotency"
    "github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * EventDispatcher
 * @desc Dispatches events to worker goroutines for parallel processing
 */
type EventDispatcher struct {
    processor   *EventProcessor
    workerCount int
    eventChan   chan *dispatchEvent
    wg          sync.WaitGroup
    ctx         context.Context
    cancel      context.CancelFunc
}

/*@
 * dispatchEvent
 * @desc Internal event wrapper
 */
type dispatchEvent struct {
    log      types.Log
    blockNum uint64
}

/*@
 * NewEventDispatcher
 * @desc Creates a new event dispatcher with worker pool
 */
func NewEventDispatcher(
    dec *decoder.EventDecoder,
    reg *handlers.Registry,
    idemp *idempotency.Tracker,
    repo *postgres.EventRepository,
    workerCount int,
) *EventDispatcher {
    if workerCount <= 0 {
        workerCount = 5
    }

    ctx, cancel := context.WithCancel(context.Background())

    return &EventDispatcher{
        processor:   NewEventProcessor(dec, reg, idemp, repo),
        workerCount: workerCount,
        eventChan:   make(chan *dispatchEvent, 1000),
        ctx:         ctx,
        cancel:      cancel,
    }
}

/*@
 * Start
 * @desc Starts the worker pool
 */
func (d *EventDispatcher) Start() error {
    for i := 0; i < d.workerCount; i++ {
        d.wg.Add(1)
        go d.worker(i)
    }
    log.Printf("Started %d event processing workers", d.workerCount)
    return nil
}

/*@
 * Stop
 * @desc Stops all workers gracefully
 */
func (d *EventDispatcher) Stop() {
    log.Println("Stopping event dispatcher...")
    d.cancel()
    d.wg.Wait()
    close(d.eventChan)
    log.Println("Event dispatcher stopped")
}

/*@
 * Dispatch
 * @desc Sends an event to the worker pool
 */
func (d *EventDispatcher) Dispatch(log types.Log, blockNum uint64) error {
    select {
    case <-d.ctx.Done():
        return d.ctx.Err()
    case d.eventChan <- &dispatchEvent{log: log, blockNum: blockNum}:
        return nil
    }
}

/*@
 * worker
 * @desc Worker goroutine that processes events
 */
func (d *EventDispatcher) worker(id int) {
    defer d.wg.Done()
    log.Printf("Worker %d started", id)

    for {
        select {
        case <-d.ctx.Done():
            log.Printf("Worker %d stopping", id)
            return
        case event, ok := <-d.eventChan:
            if !ok {
                log.Printf("Worker %d: event channel closed", id)
                return
            }
            if err := d.processor.ProcessLog(d.ctx, event.log, event.blockNum); err != nil {
                log.Printf("Worker %d: failed to process log: %v", id, err)
            }
        }
    }
}