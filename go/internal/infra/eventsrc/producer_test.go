package eventsrc

import (
	"context"
	"testing"

	"github.com/cgund98/go-eventsrc-example/internal/infra/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTransactionProducer(t *testing.T) {
	store := NewInMemoryStore()
	bus := NewInMemoryBus()
	tx := &pg.TestTransactor{}
	aggregateType := "orders"

	producer := NewTransactionProducer(store, bus, tx, aggregateType)

	assert.NotNil(t, producer)
	assert.Equal(t, store, producer.store)
	assert.Equal(t, bus, producer.bus)
	assert.Equal(t, tx, producer.tx)
	assert.Equal(t, aggregateType, producer.aggregateType)
}

func TestTransactionProducer_Send(t *testing.T) {
	store := NewInMemoryStore()
	bus := NewInMemoryBus()
	tx := &pg.TestTransactor{}
	aggregateType := "orders"

	producer := NewTransactionProducer(store, bus, tx, aggregateType)
	ctx := context.Background()

	args := &SendArgs{
		AggregateID: "order-123",
		EventType:   "OrderCreated",
		Value:       []byte(`{"amount": 100}`),
	}

	err := producer.Send(ctx, args)
	require.NoError(t, err)

	// Verify transaction was called
	assert.Equal(t, 1, tx.NumCalls)

	// Verify event was stored
	events, err := store.ListByAggregateID(ctx, args.AggregateID)
	require.NoError(t, err)
	require.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, args.AggregateID, event.AggregateID)
	assert.Equal(t, aggregateType, event.AggregateType)
	assert.Equal(t, args.EventType, event.EventType)
	assert.Equal(t, args.Value, event.Data)

	// Verify event was published to bus
	require.Len(t, bus.Events, 1)
	assert.Equal(t, args.EventType, bus.Events[0].EventType)
	assert.Equal(t, args.Value, bus.Events[0].Data)
}

func TestTransactionProducer_Send_MultipleEvents(t *testing.T) {
	store := NewInMemoryStore()
	bus := NewInMemoryBus()
	tx := &pg.TestTransactor{}
	aggregateType := "orders"

	producer := NewTransactionProducer(store, bus, tx, aggregateType)
	ctx := context.Background()

	// Send multiple events
	events := []*SendArgs{
		{
			AggregateID: "order-123",
			EventType:   "OrderCreated",
			Value:       []byte(`{"amount": 100}`),
		},
		{
			AggregateID: "order-123",
			EventType:   "OrderPaid",
			Value:       []byte(`{"payment_method": "credit_card"}`),
		},
		{
			AggregateID: "order-456",
			EventType:   "OrderCreated",
			Value:       []byte(`{"amount": 200}`),
		},
	}

	for _, event := range events {
		err := producer.Send(ctx, event)
		require.NoError(t, err)
	}

	// Verify all transactions were called
	assert.Equal(t, 3, tx.NumCalls)

	// Verify events were stored
	order123Events, err := store.ListByAggregateID(ctx, "order-123")
	require.NoError(t, err)
	assert.Len(t, order123Events, 2)

	order456Events, err := store.ListByAggregateID(ctx, "order-456")
	require.NoError(t, err)
	assert.Len(t, order456Events, 1)

	// Verify events were published to bus
	assert.Len(t, bus.Events, 3)

	// Verify the events in the bus have the correct EventType and Data
	expectedEventTypes := []string{"OrderCreated", "OrderPaid", "OrderCreated"}
	for i, expectedType := range expectedEventTypes {
		assert.Equal(t, expectedType, bus.Events[i].EventType)
	}
}
