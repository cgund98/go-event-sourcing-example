package controller

import (
	"context"
	"errors"
	"testing"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	"github.com/cgund98/go-eventsrc-example/internal/entity/orders"
	"github.com/cgund98/go-eventsrc-example/internal/infra/eventsrc"
	"github.com/cgund98/go-eventsrc-example/internal/infra/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MockStore is a mock implementation of eventsrc.Store
type MockStore struct {
	mock.Mock
}

func (m *MockStore) Persist(ctx context.Context, tx pg.Tx, args eventsrc.PersistEventArgs) (string, error) {
	callArgs := m.Called(ctx, tx, args)
	return callArgs.String(0), callArgs.Error(1)
}

func (m *MockStore) Remove(ctx context.Context, tx pg.Tx, eventId string) error {
	callArgs := m.Called(ctx, tx, eventId)
	return callArgs.Error(0)
}

func (m *MockStore) ListByAggregateID(ctx context.Context, aggregateID, aggregateType string) ([]eventsrc.Event, error) {
	callArgs := m.Called(ctx, aggregateID, aggregateType)
	return callArgs.Get(0).([]eventsrc.Event), callArgs.Error(1)
}

func createValidOrderPlacedEvent(orderId string, paymentMethod string) []byte {
	event := &pb.OrderPlaced{
		OrderId:       orderId,
		Timestamp:     timestamppb.Now(),
		VendorId:      "vendor-123",
		CustomerId:    "customer-456",
		ProductId:     "product-789",
		Quantity:      2,
		TotalPrice:    99.99,
		PaymentMethod: paymentMethod,
	}

	eventBytes, _ := proto.Marshal(event)
	return eventBytes
}

func TestController_InitializePendingPayment(t *testing.T) {
	t.Run("successful payment initialization", func(t *testing.T) {
		mockStore := &MockStore{}
		mockProducer := &MockProducer{}

		// Create valid OrderPlaced event that will produce a projection with pending status
		orderPlacedEventBytes := createValidOrderPlacedEvent("order-123", "credit_card")
		mockEvents := []eventsrc.Event{
			{
				EventType: orders.EventTypeOrderPlaced,
				Data:      orderPlacedEventBytes,
			},
		}

		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return(mockEvents, nil)
		mockProducer.On("Send", mock.Anything, mock.MatchedBy(func(args *eventsrc.SendArgs) bool {
			return args.AggregateID == "order-123" &&
				args.AggregateType == orders.AggregateTypeOrder &&
				args.EventType == orders.EventTypeOrderPaymentInitiated &&
				len(args.Value) > 0
		})).Return(nil)

		controller := &Controller{
			store:    mockStore,
			producer: mockProducer,
		}

		err := controller.InitializePendingPayment(context.Background(), "order-123")

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})

	t.Run("order not found", func(t *testing.T) {
		mockStore := &MockStore{}
		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return([]eventsrc.Event{}, nil)

		controller := &Controller{store: mockStore}

		err := controller.InitializePendingPayment(context.Background(), "order-123")

		assert.Error(t, err)
		assert.Equal(t, ErrOrderNotFound, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("get projection error", func(t *testing.T) {
		mockStore := &MockStore{}
		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return([]eventsrc.Event{}, errors.New("database error"))

		controller := &Controller{store: mockStore}

		err := controller.InitializePendingPayment(context.Background(), "order-123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list events")
		mockStore.AssertExpectations(t)
	})

	t.Run("invalid payment status", func(t *testing.T) {
		mockStore := &MockStore{}

		// Create OrderPlaced event followed by OrderPaid event to create a paid status
		orderPlacedEventBytes := createValidOrderPlacedEvent("order-123", "credit_card")
		orderPaidEvent := &pb.OrderPaid{
			OrderId:   "order-123",
			Timestamp: timestamppb.Now(),
		}
		orderPaidEventBytes, _ := proto.Marshal(orderPaidEvent)

		mockEvents := []eventsrc.Event{
			{
				EventType: orders.EventTypeOrderPlaced,
				Data:      orderPlacedEventBytes,
			},
			{
				EventType: orders.EventTypeOrderPaid,
				Data:      orderPaidEventBytes,
			},
		}

		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return(mockEvents, nil)

		controller := &Controller{store: mockStore}

		err := controller.InitializePendingPayment(context.Background(), "order-123")

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, st.Code())
		mockStore.AssertExpectations(t)
	})

	t.Run("producer send error", func(t *testing.T) {
		mockStore := &MockStore{}
		mockProducer := &MockProducer{}

		orderPlacedEventBytes := createValidOrderPlacedEvent("order-123", "credit_card")
		mockEvents := []eventsrc.Event{
			{
				EventType: orders.EventTypeOrderPlaced,
				Data:      orderPlacedEventBytes,
			},
		}

		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return(mockEvents, nil)
		mockProducer.On("Send", mock.Anything, mock.Anything).Return(errors.New("kafka error"))

		controller := &Controller{
			store:    mockStore,
			producer: mockProducer,
		}

		err := controller.InitializePendingPayment(context.Background(), "order-123")

		assert.Error(t, err)
		assert.Equal(t, ErrInternal, err)
		mockStore.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})
}

func TestValidatePaymentRequest(t *testing.T) {
	t.Run("valid pending status", func(t *testing.T) {
		projection := &orders.OrderProjection{
			PaymentMethod: "credit_card",
			PaymentStatus: orders.PaymentStatusPending,
		}

		err := validateInitPendingPaymentRequest(projection)
		assert.NoError(t, err)
	})

	t.Run("missing payment method", func(t *testing.T) {
		projection := &orders.OrderProjection{
			PaymentMethod: "",
			PaymentStatus: orders.PaymentStatusPending,
		}

		err := validateInitPendingPaymentRequest(projection)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("invalid payment status", func(t *testing.T) {
		projection := &orders.OrderProjection{
			PaymentMethod: "credit_card",
			PaymentStatus: orders.PaymentStatusPaid,
		}

		err := validateInitPendingPaymentRequest(projection)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, st.Code())
	})
}
