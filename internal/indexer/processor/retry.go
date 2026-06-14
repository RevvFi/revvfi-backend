// internal/indexer/processor/retry.go
package processor

import (
    "context"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "time"

    "github.com/ethereum/go-ethereum/common"
    goEthtypes "github.com/ethereum/go-ethereum/core/types"

    "github.com/Revvfi/revvfi-backend/internal/indexer/decoder"
    "github.com/Revvfi/revvfi-backend/internal/handlers"
    indexertypes "github.com/Revvfi/revvfi-backend/internal/indexer/types"
    "github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * FailedEventProcessor
 * @desc Processes failed events from the dead letter queue
 */
type FailedEventProcessor struct {
    eventRepo  *postgres.EventRepository
    decoder    *decoder.EventDecoder
    handler    *handlers.Registry
    maxRetries int
    backoff    time.Duration
    maxBackoff time.Duration
}

/*@
 * NewFailedEventProcessor
 * @desc Creates a new failed event processor
 */
func NewFailedEventProcessor(
    repo *postgres.EventRepository,
    dec *decoder.EventDecoder,
    h *handlers.Registry,
) *FailedEventProcessor {
    return &FailedEventProcessor{
        eventRepo:  repo,
        decoder:    dec,
        handler:    h,
        maxRetries: 5,
        backoff:    2 * time.Second,
        maxBackoff: 5 * time.Minute,
    }
}

/*@
 * ProcessFailedEvents
 * @desc Main method to process all pending failed events
 */
func (f *FailedEventProcessor) ProcessFailedEvents(ctx context.Context) error {
    // Fetch events ready for retry (up to 100 per batch)
    events, err := f.eventRepo.GetUnresolvedFailedEvents(ctx, 100)
    if err != nil {
        return fmt.Errorf("failed to fetch failed events: %w", err)
    }

    if len(events) == 0 {
        return nil
    }

    for _, ev := range events {
        // Check if max retries exceeded
        if ev.RetryCount >= f.maxRetries {
            // Mark as permanently failed - no more retries
            if err := f.eventRepo.MarkFailedEventResolved(ctx, ev.ID, false); err != nil {
                // Log error but continue
                continue
            }
            continue
        }

        // Reconstruct raw log from stored data
        rawLog, err := f.reconstructRawLog(ev)
        if err != nil {
            f.eventRepo.MarkFailedEventResolved(ctx, ev.ID, false)
            continue
        }

        // Decode the event
        decodedEvent, err := f.decoder.Decode(*rawLog)
        if err != nil {
            f.eventRepo.IncrementFailedEventRetry(ctx, ev.ID, fmt.Sprintf("decode error: %v", err))
            continue
        }

        // Get the appropriate handler for this event
        eventName := decoder.GetEventName(rawLog.Topics[0].Hex())
        handler, ok := f.handler.Get(eventName)
        if !ok {
            f.eventRepo.MarkFailedEventResolved(ctx, ev.ID, false)
            continue
        }

        // Process the event
        if err := handler.Handle(ctx, decodedEvent, rawLog.BlockNumber); err != nil {
            f.eventRepo.IncrementFailedEventRetry(ctx, ev.ID, fmt.Sprintf("handler error: %v", err))
            continue
        }

        // Success! Mark as resolved
        if err := f.eventRepo.MarkFailedEventResolved(ctx, ev.ID, true); err != nil {
            continue
        }
    }

    return nil
}

/*@
 * calculateBackoff
 * @desc Calculates exponential backoff duration
 */
func (f *FailedEventProcessor) calculateBackoff(retryCount int) time.Duration {
    backoff := f.backoff * time.Duration(1<<retryCount)
    if backoff > f.maxBackoff {
        backoff = f.maxBackoff
    }
    return backoff
}

/*@
 * reconstructRawLog
 * @desc Reconstructs a types.Log from stored failed event data
 */
func (f *FailedEventProcessor) reconstructRawLog(event *postgres.FailedEventRecord) (*goEthtypes.Log, error) {
    // Parse event data as LogData
    var logData indexertypes.LogData
    if err := json.Unmarshal(event.EventData, &logData); err != nil {
        return nil, fmt.Errorf("failed to unmarshal log data: %w", err)
    }

    // Convert topics from strings to common.Hash
    topics := make([]common.Hash, len(logData.Topics))
    for i, topicStr := range logData.Topics {
        topics[i] = common.HexToHash(topicStr)
    }

    // Decode data from hex string to bytes
    data, err := hex.DecodeString(logData.Data[2:])
    if err != nil {
        return nil, fmt.Errorf("failed to decode log data: %w", err)
    }

    // Reconstruct the raw log
    rawLog := &goEthtypes.Log{
        Address:     common.HexToAddress(logData.Address),
        Topics:      topics,
        Data:        data,
        BlockNumber: logData.BlockNumber,
        TxHash:      common.HexToHash(logData.TxHash),
        TxIndex:     logData.TxIndex,
        BlockHash:   common.HexToHash(logData.BlockHash),
        Index:       logData.Index,
        Removed:     logData.Removed,
    }

    return rawLog, nil
}

/*@
 * StartRetryWorker
 * @desc Starts a background worker that periodically processes failed events
 */
func (f *FailedEventProcessor) StartRetryWorker(ctx context.Context, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            _ = f.ProcessFailedEvents(ctx)
        }
    }
}