package eventsrc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryStore_Persist(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	args := PersistEventArgs{
		AggregateId:   "order-123",
		AggregateType: "Order",
		EventType:     "OrderCreated",
		Data:          []byte(`{"amount": 100}`),
	}

	err := store.Persist(ctx, nil, args)
	require.NoError(t, err)

	// Verify event was stored
	events, err := store.ListByAggregateID(ctx, args.AggregateId, args.AggregateType)
	require.NoError(t, err)
	require.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, 0, event.EventId)
	assert.Equal(t, args.AggregateId, event.AggregateId)
	assert.Equal(t, args.AggregateType, event.AggregateType)
	assert.Equal(t, args.EventType, event.EventType)
	assert.Equal(t, args.Data, event.Data)
	assert.WithinDuration(t, time.Now().UTC(), event.CreatedAt, 2*time.Second)
}

func TestInMemoryStore_MultipleEvents(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()
	aggregateID := "order-123"

	// Persist multiple events
	events := []PersistEventArgs{
		{
			AggregateId:   aggregateID,
			AggregateType: "Order",
			EventType:     "OrderCreated",
			Data:          []byte(`{"amount": 100}`),
		},
		{
			AggregateId:   aggregateID,
			AggregateType: "Order",
			EventType:     "OrderPaid",
			Data:          []byte(`{"payment_method": "credit_card"}`),
		},
	}

	for _, event := range events {
		err := store.Persist(ctx, nil, event)
		require.NoError(t, err)
	}

	// Verify all events are stored
	storedEvents, err := store.ListByAggregateID(ctx, aggregateID, "Order")
	require.NoError(t, err)
	assert.Len(t, storedEvents, 2)

	// Verify event IDs are sequential
	for i, event := range storedEvents {
		assert.Equal(t, i, event.EventId)
		assert.Equal(t, aggregateID, event.AggregateId)
	}
}

func TestInMemoryStore_ListByAggregateID_Empty(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	events, err := store.ListByAggregateID(ctx, "non-existent", "Order")
	require.NoError(t, err)
	assert.Empty(t, events)
}
