package eventsrc

import (
	"context"
	"database/sql"

	"github.com/cgund98/go-eventsrc-example/internal/infra/logging"
	"github.com/cgund98/go-eventsrc-example/internal/infra/pg"
)

// SendArgs contains the arguments required to send an event.
type SendArgs struct {
	AggregateID   string
	AggregateType string
	EventType     string
	Value         []byte
}

// Producer is the interface for sending events.
type Producer interface {
	Send(ctx context.Context, args *SendArgs) error
}

// TransactionProducer implements Producer and handles transactional event sending.
type TransactionProducer struct {
	store Store
	bus   Bus
	tx    pg.Transactor
}

// NewTransactionProducer creates a new TransactionProducer.
func NewTransactionProducer(store Store, bus Bus, tx pg.Transactor) *TransactionProducer {
	return &TransactionProducer{store: store, bus: bus, tx: tx}
}

// Send sends an event transactionally using the configured store and bus.
func (p *TransactionProducer) Send(ctx context.Context, args *SendArgs) error {
	var eventId string

	// We need to commit our event to the store before publishing it to the bus.
	// Otherwise the event may be consumed before it is committed to the store.
	err := p.tx.WithTx(ctx, &sql.TxOptions{}, func(tx pg.Tx) error {
		addedEventId, err := p.store.Persist(ctx, tx, PersistEventArgs{
			AggregateId:   args.AggregateID,
			AggregateType: args.AggregateType,
			EventType:     args.EventType,
			Data:          args.Value,
		})
		if err != nil {
			return err
		}
		eventId = addedEventId
		return nil
	})
	if err != nil {
		return err
	}

	err = p.bus.Publish(ctx, &PublishArgs{
		AggregateID:   args.AggregateID,
		AggregateType: args.AggregateType,
		EventType:     args.EventType,
		Value:         args.Value,
	})

	// If the event is not published, remove it from the store and return an error.
	// This is to avoid a race condition where the event is consumed before it is committed to the store.
	if err != nil {
		logging.Logger.Info("failed to publish event. attempting to remove it from the store", "error", err)
		removeErr := p.tx.WithTx(ctx, &sql.TxOptions{}, func(tx pg.Tx) error {
			return p.store.Remove(ctx, tx, eventId)
		})
		if removeErr != nil {
			logging.Logger.Error("failed to remove event from store", "error", removeErr)
		}

		return err
	}

	return nil
}
