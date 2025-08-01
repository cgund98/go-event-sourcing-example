package eventsrc

import (
	"context"
	"database/sql"

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
	return p.tx.WithTx(ctx, &sql.TxOptions{}, func(tx pg.Tx) error {
		err := p.store.Persist(ctx, tx, PersistEventArgs{
			AggregateId:   args.AggregateID,
			AggregateType: args.AggregateType,
			EventType:     args.EventType,
			Data:          args.Value,
		})
		if err != nil {
			return err
		}

		return p.bus.Publish(ctx, &PublishArgs{
			EventType: args.EventType,
			Value:     args.Value,
		})
	})
}
