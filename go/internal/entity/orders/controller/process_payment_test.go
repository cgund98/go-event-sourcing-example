package controller

import (
	"context"
	"errors"
	"testing"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	"github.com/cgund98/go-eventsrc-example/internal/entity/orders"
	"github.com/cgund98/go-eventsrc-example/internal/infra/eventsrc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func createValidOrderPaymentInitiatedEvent(orderId string) []byte {
	event := &pb.OrderPaymentInitiated{
		OrderId:   orderId,
		Timestamp: timestamppb.Now(),
	}

	eventBytes, _ := proto.Marshal(event)
	return eventBytes
}

func TestController_ProcessPayment(t *testing.T) {
	t.Run("successful payment processing", func(t *testing.T) {
		mockStore := &MockStore{}
		mockProducer := &MockProducer{}

		// Create events: OrderPlaced -> OrderPaymentInitiated (creates initiated status)
		orderPlacedEventBytes := createValidOrderPlacedEvent("order-123", "credit_card")
		orderPaymentInitiatedEventBytes := createValidOrderPaymentInitiatedEvent("order-123")

		mockEvents := []eventsrc.Event{
			{
				EventType:      orders.EventTypeOrderPlaced,
				Data:           orderPlacedEventBytes,
				SequenceNumber: 0,
			},
			{
				EventType:      orders.EventTypeOrderPaymentInitiated,
				Data:           orderPaymentInitiatedEventBytes,
				SequenceNumber: 1,
			},
		}

		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return(mockEvents, nil)
		mockProducer.On("Send", mock.Anything, mock.MatchedBy(func(args *eventsrc.SendArgs) bool {
			return args.AggregateID == "order-123" &&
				args.AggregateType == orders.AggregateTypeOrder &&
				args.EventType == orders.EventTypeOrderPaid &&
				args.SequenceNumber == 2 &&
				len(args.Value) > 0
		})).Return(nil)

		controller := &Controller{
			store:    mockStore,
			producer: mockProducer,
		}

		err := controller.ProcessPayment(context.Background(), "order-123")

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})

	t.Run("order not found", func(t *testing.T) {
		mockStore := &MockStore{}
		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return([]eventsrc.Event{}, nil)

		controller := &Controller{store: mockStore}

		err := controller.ProcessPayment(context.Background(), "order-123")

		assert.Error(t, err)
		assert.Equal(t, ErrOrderNotFound, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("get projection error", func(t *testing.T) {
		mockStore := &MockStore{}
		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return([]eventsrc.Event{}, errors.New("database error"))

		controller := &Controller{store: mockStore}

		err := controller.ProcessPayment(context.Background(), "order-123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list events")
		mockStore.AssertExpectations(t)
	})

	t.Run("invalid payment status", func(t *testing.T) {
		mockStore := &MockStore{}

		// Create OrderPlaced event only (creates pending status, not initiated)
		orderPlacedEventBytes := createValidOrderPlacedEvent("order-123", "credit_card")
		mockEvents := []eventsrc.Event{
			{
				EventType:      orders.EventTypeOrderPlaced,
				Data:           orderPlacedEventBytes,
				SequenceNumber: 0,
			},
		}

		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return(mockEvents, nil)

		controller := &Controller{store: mockStore}

		err := controller.ProcessPayment(context.Background(), "order-123")

		assert.Error(t, err)
		assert.Equal(t, ErrPaymentStatusNotInitiated, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("missing payment method", func(t *testing.T) {
		mockStore := &MockStore{}

		// Create events: OrderPlaced (no payment method) -> OrderPaymentInitiated
		orderPlacedEventBytes := createValidOrderPlacedEvent("order-123", "") // No payment method
		orderPaymentInitiatedEventBytes := createValidOrderPaymentInitiatedEvent("order-123")

		mockEvents := []eventsrc.Event{
			{
				EventType:      orders.EventTypeOrderPlaced,
				Data:           orderPlacedEventBytes,
				SequenceNumber: 0,
			},
			{
				EventType:      orders.EventTypeOrderPaymentInitiated,
				Data:           orderPaymentInitiatedEventBytes,
				SequenceNumber: 1,
			},
		}

		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return(mockEvents, nil)

		controller := &Controller{store: mockStore}

		err := controller.ProcessPayment(context.Background(), "order-123")

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "payment method is required")
		mockStore.AssertExpectations(t)
	})

	t.Run("producer send error", func(t *testing.T) {
		mockStore := &MockStore{}
		mockProducer := &MockProducer{}

		orderPlacedEventBytes := createValidOrderPlacedEvent("order-123", "credit_card")
		orderPaymentInitiatedEventBytes := createValidOrderPaymentInitiatedEvent("order-123")

		mockEvents := []eventsrc.Event{
			{
				EventType:      orders.EventTypeOrderPlaced,
				Data:           orderPlacedEventBytes,
				SequenceNumber: 0,
			},
			{
				EventType:      orders.EventTypeOrderPaymentInitiated,
				Data:           orderPaymentInitiatedEventBytes,
				SequenceNumber: 1,
			},
		}

		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return(mockEvents, nil)
		mockProducer.On("Send", mock.Anything, mock.Anything).Return(errors.New("kafka error"))

		controller := &Controller{
			store:    mockStore,
			producer: mockProducer,
		}

		err := controller.ProcessPayment(context.Background(), "order-123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send event")
		mockStore.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})
}

func TestValidateProcessPaymentRequest(t *testing.T) {
	t.Run("valid initiated status", func(t *testing.T) {
		projection := &orders.OrderProjection{
			PaymentMethod: "credit_card",
			PaymentStatus: orders.PaymentStatusInitiated,
		}

		err := validateProcessPaymentRequest(projection)
		assert.NoError(t, err)
	})

	t.Run("missing payment method", func(t *testing.T) {
		projection := &orders.OrderProjection{
			PaymentMethod: "",
			PaymentStatus: orders.PaymentStatusInitiated,
		}

		err := validateProcessPaymentRequest(projection)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "payment method is required")
	})

	t.Run("invalid payment status", func(t *testing.T) {
		projection := &orders.OrderProjection{
			PaymentMethod: "credit_card",
			PaymentStatus: orders.PaymentStatusPending,
		}

		err := validateProcessPaymentRequest(projection)
		assert.Error(t, err)
		assert.Equal(t, ErrPaymentStatusNotInitiated, err)
	})
}
