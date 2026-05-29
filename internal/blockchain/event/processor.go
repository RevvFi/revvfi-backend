package event

import (
	"context"

	"github.com/Revvfi/revvfi-backend/internal/blockchain/types"
	"github.com/Revvfi/revvfi-backend/internal/models"
	"github.com/Revvfi/revvfi-backend/internal/repository"
)

/*
@struct Processor

@desc
Processes normalized blockchain events with idempotency protection.

@responsibilities
- Check event processing status
- Mark successful event processing
- Provide a central point for future domain-specific mutations
*/
type Processor struct {
	events repository.EventRepository
}

/*
@function NewProcessor

@desc
Creates a blockchain event processor.

@params
- events: event idempotency repository

@returns
- *Processor
*/
func NewProcessor(events repository.EventRepository) *Processor {
	return &Processor{events: events}
}

/*
@method Process

@desc
Processes one normalized blockchain event.

@params
- ctx: processing lifecycle context
- event: normalized event payload

@returns
- error
*/
func (p *Processor) Process(ctx context.Context, event types.Event) error {
	processed, err := p.events.IsProcessed(ctx, event.TxHash, event.LogIndex)
	if err != nil {
		return err
	}
	if processed {
		return nil
	}

	return p.events.MarkProcessed(ctx, &models.ProcessedEvent{
		TxHash:          event.TxHash,
		LogIndex:        event.LogIndex,
		BlockNumber:     event.BlockNumber,
		EventName:       event.Name,
		ContractAddress: event.ContractAddress,
	})
}
