// internal/indexer/processor/processor.go
package processor

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Revvfi/revvfi-backend/internal/handlers"
	"github.com/Revvfi/revvfi-backend/internal/indexer/decoder"
	"github.com/Revvfi/revvfi-backend/internal/indexer/idempotency"
	"github.com/Revvfi/revvfi-backend/internal/repository/postgres"
)

/*@
 * EventProcessor
 * @desc Core event processing engine
 * @responsibilities
 *   - Receive raw blockchain logs
 *   - Decode events using ABI
 *   - Check idempotency (prevent duplicates)
 *   - Route to appropriate handler
 *   - Track metrics
 *   - Handle failures
 *
 * @flow
 *   Raw Log → Decode → Idempotency Check → Handler → Database → Mark Processed
 */
type EventProcessor struct {
	decoder     *decoder.EventDecoder
	handlerReg  *handlers.Registry
	idempotency *idempotency.Tracker
	eventRepo   *postgres.EventRepository
}

/*@
 * NewEventProcessor
 * @desc Creates a new event processor instance
 * @param dec Event decoder for converting raw logs to typed events
 * @param reg Handler registry for routing events
 * @param idemp Idempotency tracker for duplicate prevention
 * @param repo Event repository for database operations
 * @returns Configured EventProcessor
 */
func NewEventProcessor(
	dec *decoder.EventDecoder,
	reg *handlers.Registry,
	idemp *idempotency.Tracker,
	repo *postgres.EventRepository,
) *EventProcessor {
	return &EventProcessor{
		decoder:     dec,
		handlerReg:  reg,
		idempotency: idemp,
		eventRepo:   repo,
	}
}

/*@
 * ProcessLog
 * @desc Processes a single raw blockchain log
 * @param ctx Context for cancellation
 * @param log Raw Ethereum log from RPC
 * @param blockNum Block number where log occurred
 * @returns error if processing fails
 *
 * @flow
 *   1. Get event name from topic
 *   2. Check if already processed (idempotency)
 *   3. Decode log to typed event
 *   4. Get appropriate handler
 *   5. Execute handler
 *   6. Mark as processed
 */
func (p *EventProcessor) ProcessLog(ctx context.Context, logEntry types.Log, blockNum uint64) error {
	// Get event name from first topic
	if len(logEntry.Topics) == 0 {
		return nil
	}
	
	eventName := decoder.GetEventName(logEntry.Topics[0].Hex())
	log.Printf(
	"Topic=%s Event=%s",
	logEntry.Topics[0].Hex(),
	eventName,
)
	if eventName == "" {
		// Unknown event - skip (not a RevvFi event)
		return nil
	}

	// Check idempotency (prevent duplicate processing)
	if processed, err := p.idempotency.IsProcessed(ctx, logEntry.TxHash.Hex(), int(logEntry.Index)); err != nil {
		log.Printf("Failed to check idempotency for tx=%s: %v", logEntry.TxHash.Hex(), err)
		return err
	} else if processed {
		// Already processed, skip
		log.Printf("Event already processed: %s (tx=%s, logIndex=%d)", eventName, logEntry.TxHash.Hex(), logEntry.Index)
		return nil
	}

	// Decode the event
	decodedEvent, err := p.decoder.Decode(logEntry)
	if err != nil {
		log.Printf("Failed to decode event %s: %v", eventName, err)
		// Store failed event for retry
		p.storeFailedEvent(ctx, eventName, logEntry, err.Error())
		return err
	}

	// Get the handler for this event
	handler, ok := p.handlerReg.Get(eventName)
	if !ok {
		log.Printf("No handler registered for event: %s", eventName)
		// Store as failed - no handler available
		p.storeFailedEvent(ctx, eventName, logEntry, "no handler registered")
		return nil
	}

	// Process the event
	if err := handler.Handle(ctx, decodedEvent, blockNum); err != nil {
		log.Printf("Handler failed for event %s: %v", eventName, err)
		// Store failed event for retry
		p.storeFailedEvent(ctx, eventName, logEntry, err.Error())
		return err
	}

	// Mark as processed (idempotency)
	if err := p.idempotency.MarkProcessed(ctx, logEntry.TxHash.Hex(), int(logEntry.Index), blockNum, eventName, logEntry.Address.Hex()); err != nil {
		log.Printf("Failed to mark event as processed: %v", err)
		return err
	}

	log.Printf("Successfully processed event: %s (tx=%s, block=%d)", eventName, logEntry.TxHash.Hex(), blockNum)
	return nil
}

/*@
 * storeFailedEvent
 * @desc Stores failed event in dead letter queue for later retry
 * @param ctx Context for cancellation
 * @param eventName Name of the event
 * @param logEntry Raw blockchain log
 * @param reason Failure reason
 */
func (p *EventProcessor) storeFailedEvent(ctx context.Context, eventName string, logEntry types.Log, reason string) {
	// Convert log to serializable format
	logData := map[string]interface{}{
		"address":     logEntry.Address.Hex(),
		"topics":      topicsToHex(logEntry.Topics),
		"data":        common.Bytes2Hex(logEntry.Data),
		"blockNumber": logEntry.BlockNumber,
		"txHash":      logEntry.TxHash.Hex(),
		"txIndex":     logEntry.TxIndex,
		"blockHash":   logEntry.BlockHash.Hex(),
		"logIndex":    logEntry.Index,
		"removed":     logEntry.Removed,
	}

	if err := p.eventRepo.StoreFailedEvent(ctx, eventName, logData, nil, reason, 5); err != nil {
		log.Printf("Failed to store failed event: %v", err)
	}
}

/*@
 * topicsToHex
 * @desc Converts slice of common.Hash to slice of hex strings
 * @param topics Slice of topic hashes
 * @returns Slice of hex strings
 */
func topicsToHex(topics []common.Hash) []string {
	result := make([]string, len(topics))
	for i, topic := range topics {
		result[i] = topic.Hex()
	}
	return result
}
