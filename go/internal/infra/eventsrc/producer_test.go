package eventsrc

import (
	"context"
	"errors"
	"testing"

	"github.com/cgund98/go-eventsrc-example/internal/infra/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockBus is a mock implementation of Bus that can be configured to fail
type MockBus struct {
	mock.Mock
}

func (m *MockBus) Publish(ctx context.Context, args *PublishArgs) error {
	callArgs := m.Called(ctx, args)
	return callArgs.Error(0)
}

// MockStore is a mock implementation of Store that can track operations
type MockStore struct {
	mock.Mock
	persistedEvents []Event
	removedEventIds []string
}

func (m *MockStore) Persist(ctx context.Context, tx pg.Tx, args PersistEventArgs) (string, error) {
	callArgs := m.Called(ctx, tx, args)
	eventId := callArgs.String(0)
	if callArgs.Error(1) == nil {
		m.persistedEvents = append(m.persistedEvents, Event{
			EventId:       1, // Mock event ID
			AggregateId:   args.AggregateId,
			AggregateType: args.AggregateType,
			EventType:     args.EventType,
			Data:          args.Data,
		})
	}
	return eventId, callArgs.Error(1)
}

func (m *MockStore) Remove(ctx context.Context, tx pg.Tx, eventId string) error {
	callArgs := m.Called(ctx, tx, eventId)
	if callArgs.Error(0) == nil {
		m.removedEventIds = append(m.removedEventIds, eventId)
	}
	return callArgs.Error(0)
}

func (m *MockStore) ListByAggregateID(ctx context.Context, aggregateId string, aggregateType string) ([]Event, error) {
	callArgs := m.Called(ctx, aggregateId, aggregateType)
	return callArgs.Get(0).([]Event), callArgs.Error(1)
}

func TestNewTransactionProducer(t *testing.T) {
	store := NewInMemoryStore()
	bus := NewInMemoryBus()
	tx := &pg.TestTransactor{}

	producer := NewTransactionProducer(store, bus, tx)

	assert.NotNil(t, producer)
	assert.Equal(t, store, producer.store)
	assert.Equal(t, bus, producer.bus)
	assert.Equal(t, tx, producer.tx)
}

func TestTransactionProducer_Send(t *testing.T) {
	store := NewInMemoryStore()
	bus := NewInMemoryBus()
	tx := &pg.TestTransactor{}

	producer := NewTransactionProducer(store, bus, tx)
	ctx := context.Background()

	args := &SendArgs{
		AggregateID:   "order-123",
		AggregateType: "orders",
		EventType:     "OrderCreated",
		Value:         []byte(`{"amount": 100}`),
	}

	err := producer.Send(ctx, args)
	require.NoError(t, err)

	// Verify transaction was called
	assert.Equal(t, 1, tx.NumCalls)

	// Verify event was stored
	events, err := store.ListByAggregateID(ctx, args.AggregateID, args.AggregateType)
	require.NoError(t, err)
	require.Len(t, events, 1)

	event := events[0]
	assert.Equal(t, args.AggregateID, event.AggregateId)
	assert.Equal(t, args.AggregateType, event.AggregateType)
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

	producer := NewTransactionProducer(store, bus, tx)
	ctx := context.Background()

	// Send multiple events
	events := []*SendArgs{
		{
			AggregateID:   "order-123",
			AggregateType: "orders",
			EventType:     "OrderCreated",
			Value:         []byte(`{"amount": 100}`),
		},
		{
			AggregateID:   "order-123",
			AggregateType: "orders",
			EventType:     "OrderPaid",
			Value:         []byte(`{"payment_method": "credit_card"}`),
		},
		{
			AggregateID:   "order-456",
			AggregateType: "orders",
			EventType:     "OrderCreated",
			Value:         []byte(`{"amount": 200}`),
		},
	}

	for _, event := range events {
		err := producer.Send(ctx, event)
		require.NoError(t, err)
	}

	// Verify all transactions were called
	assert.Equal(t, 3, tx.NumCalls)

	// Verify events were stored
	order123Events, err := store.ListByAggregateID(ctx, "order-123", "orders")
	require.NoError(t, err)
	assert.Len(t, order123Events, 2)

	order456Events, err := store.ListByAggregateID(ctx, "order-456", "orders")
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

func TestTransactionProducer_Send_BusPublishFailure(t *testing.T) {
	mockStore := &MockStore{}
	mockBus := &MockBus{}
	tx := &pg.TestTransactor{}

	producer := NewTransactionProducer(mockStore, mockBus, tx)
	ctx := context.Background()

	args := &SendArgs{
		AggregateID:   "order-123",
		AggregateType: "orders",
		EventType:     "OrderCreated",
		Value:         []byte(`{"amount": 100}`),
	}

	// Configure mocks
	mockStore.On("Persist", mock.Anything, mock.Anything, mock.MatchedBy(func(persistArgs PersistEventArgs) bool {
		return persistArgs.AggregateId == args.AggregateID &&
			persistArgs.AggregateType == args.AggregateType &&
			persistArgs.EventType == args.EventType &&
			string(persistArgs.Data) == string(args.Value)
	})).Return("event-123", nil)

	mockBus.On("Publish", mock.Anything, mock.MatchedBy(func(publishArgs *PublishArgs) bool {
		return publishArgs.EventType == args.EventType &&
			string(publishArgs.Value) == string(args.Value)
	})).Return(errors.New("kafka connection failed"))

	mockStore.On("Remove", mock.Anything, mock.Anything, "event-123").Return(nil)

	// Execute Send
	err := producer.Send(ctx, args)

	// Verify error is returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kafka connection failed")

	// Verify Persist was called
	mockStore.AssertCalled(t, "Persist", mock.Anything, mock.Anything, mock.Anything)

	// Verify Publish was called
	mockBus.AssertCalled(t, "Publish", mock.Anything, mock.Anything)

	// Verify Remove was called to rollback
	mockStore.AssertCalled(t, "Remove", mock.Anything, mock.Anything, "event-123")

	// Verify the event was actually removed
	assert.Contains(t, mockStore.removedEventIds, "event-123")
}

func TestTransactionProducer_Send_BusPublishFailure_RemoveError(t *testing.T) {
	mockStore := &MockStore{}
	mockBus := &MockBus{}
	tx := &pg.TestTransactor{}

	producer := NewTransactionProducer(mockStore, mockBus, tx)
	ctx := context.Background()

	args := &SendArgs{
		AggregateID:   "order-123",
		AggregateType: "orders",
		EventType:     "OrderCreated",
		Value:         []byte(`{"amount": 100}`),
	}

	// Configure mocks
	mockStore.On("Persist", mock.Anything, mock.Anything, mock.Anything).Return("event-123", nil)
	mockBus.On("Publish", mock.Anything, mock.Anything).Return(errors.New("kafka connection failed"))
	mockStore.On("Remove", mock.Anything, mock.Anything, "event-123").Return(errors.New("database error"))

	// Execute Send
	err := producer.Send(ctx, args)

	// Verify error is returned (should be the original publish error, not the remove error)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kafka connection failed")

	// Verify all operations were called
	mockStore.AssertCalled(t, "Persist", mock.Anything, mock.Anything, mock.Anything)
	mockBus.AssertCalled(t, "Publish", mock.Anything, mock.Anything)
	mockStore.AssertCalled(t, "Remove", mock.Anything, mock.Anything, "event-123")
}

func TestTransactionProducer_Send_StorePersistFailure(t *testing.T) {
	mockStore := &MockStore{}
	mockBus := &MockBus{}
	tx := &pg.TestTransactor{}

	producer := NewTransactionProducer(mockStore, mockBus, tx)
	ctx := context.Background()

	args := &SendArgs{
		AggregateID:   "order-123",
		AggregateType: "orders",
		EventType:     "OrderCreated",
		Value:         []byte(`{"amount": 100}`),
	}

	// Configure mocks - Persist fails
	mockStore.On("Persist", mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("database error"))

	// Execute Send
	err := producer.Send(ctx, args)

	// Verify error is returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")

	// Verify Persist was called
	mockStore.AssertCalled(t, "Persist", mock.Anything, mock.Anything, mock.Anything)

	// Verify Publish was NOT called (since Persist failed)
	mockBus.AssertNotCalled(t, "Publish")

	// Verify Remove was NOT called (since Persist failed)
	mockStore.AssertNotCalled(t, "Remove")
}
