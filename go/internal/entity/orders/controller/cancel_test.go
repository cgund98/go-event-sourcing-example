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

func TestController_CancelOrder(t *testing.T) {
	t.Run("successful order cancellation", func(t *testing.T) {
		mockStore := &MockStore{}
		mockProducer := &MockProducer{}

		// Create events: OrderPlaced -> OrderPaid (creates paid status)
		orderPlacedEventBytes := createValidOrderPlacedEvent("order-123", "credit_card")
		orderPaidEventBytes := createValidOrderPaidEvent("order-123")

		mockEvents := []eventsrc.Event{
			{
				EventType:      orders.EventTypeOrderPlaced,
				Data:           orderPlacedEventBytes,
				SequenceNumber: 0,
			},
			{
				EventType:      orders.EventTypeOrderPaid,
				Data:           orderPaidEventBytes,
				SequenceNumber: 1,
			},
		}

		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return(mockEvents, nil)
		mockProducer.On("Send", mock.Anything, mock.MatchedBy(func(args *eventsrc.SendArgs) bool {
			return args.AggregateID == "order-123" &&
				args.AggregateType == orders.AggregateTypeOrder &&
				args.EventType == orders.EventTypeOrderCancelled &&
				args.SequenceNumber == 2 &&
				len(args.Value) > 0
		})).Return(nil)

		controller := &Controller{
			store:    mockStore,
			producer: mockProducer,
		}

		request := &pb.CancelOrderRequest{
			OrderId: "order-123",
			Reason:  "Customer requested cancellation",
		}

		response, err := controller.CancelOrder(context.Background(), request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "order-123", response.OrderId)
		mockStore.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})

	t.Run("order not found", func(t *testing.T) {
		mockStore := &MockStore{}
		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return([]eventsrc.Event{}, nil)

		controller := &Controller{store: mockStore}

		request := &pb.CancelOrderRequest{
			OrderId: "order-123",
			Reason:  "Customer requested cancellation",
		}

		response, err := controller.CancelOrder(context.Background(), request)

		assert.Error(t, err)
		assert.Equal(t, ErrOrderNotFound, err)
		assert.Nil(t, response)
		mockStore.AssertExpectations(t)
	})

	t.Run("order already cancelled", func(t *testing.T) {
		mockStore := &MockStore{}

		// Create events: OrderPlaced -> OrderPaid -> OrderCancelled
		orderPlacedEventBytes := createValidOrderPlacedEvent("order-123", "credit_card")
		orderPaidEventBytes := createValidOrderPaidEvent("order-123")

		orderCancelledEvent := &pb.OrderCancelled{
			OrderId:   "order-123",
			Timestamp: timestamppb.Now(),
			Reason:    "Previous cancellation",
		}
		orderCancelledEventBytes, _ := proto.Marshal(orderCancelledEvent)

		mockEvents := []eventsrc.Event{
			{
				EventType:      orders.EventTypeOrderPlaced,
				Data:           orderPlacedEventBytes,
				SequenceNumber: 0,
			},
			{
				EventType:      orders.EventTypeOrderPaid,
				Data:           orderPaidEventBytes,
				SequenceNumber: 1,
			},
			{
				EventType:      orders.EventTypeOrderCancelled,
				Data:           orderCancelledEventBytes,
				SequenceNumber: 2,
			},
		}

		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return(mockEvents, nil)

		controller := &Controller{store: mockStore}

		request := &pb.CancelOrderRequest{
			OrderId: "order-123",
			Reason:  "Customer requested cancellation",
		}

		response, err := controller.CancelOrder(context.Background(), request)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, st.Code())
		assert.Contains(t, st.Message(), "order is already cancelled")
		assert.Nil(t, response)
		mockStore.AssertExpectations(t)
	})

	t.Run("order already delivered", func(t *testing.T) {
		mockStore := &MockStore{}

		// Create events: OrderPlaced -> OrderPaid -> ShippingStatusUpdated (delivered)
		orderPlacedEventBytes := createValidOrderPlacedEvent("order-123", "credit_card")
		orderPaidEventBytes := createValidOrderPaidEvent("order-123")

		shippingEvent := &pb.OrderShippingStatusUpdated{
			OrderId:   "order-123",
			Timestamp: timestamppb.Now(),
			Status:    pb.ShippingStatus_SHIPPING_STATUS_DELIVERED,
		}
		shippingEventBytes, _ := proto.Marshal(shippingEvent)

		mockEvents := []eventsrc.Event{
			{
				EventType: orders.EventTypeOrderPlaced,
				Data:      orderPlacedEventBytes,
			},
			{
				EventType: orders.EventTypeOrderPaid,
				Data:      orderPaidEventBytes,
			},
			{
				EventType: orders.EventTypeOrderShippingStatusUpdated,
				Data:      shippingEventBytes,
			},
		}

		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return(mockEvents, nil)

		controller := &Controller{store: mockStore}

		request := &pb.CancelOrderRequest{
			OrderId: "order-123",
			Reason:  "Customer requested cancellation",
		}

		response, err := controller.CancelOrder(context.Background(), request)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, st.Code())
		assert.Contains(t, st.Message(), "cannot cancel an order that has already been delivered")
		assert.Nil(t, response)
		mockStore.AssertExpectations(t)
	})

	t.Run("producer send error", func(t *testing.T) {
		mockStore := &MockStore{}
		mockProducer := &MockProducer{}

		orderPlacedEventBytes := createValidOrderPlacedEvent("order-123", "credit_card")
		orderPaidEventBytes := createValidOrderPaidEvent("order-123")

		mockEvents := []eventsrc.Event{
			{
				EventType:      orders.EventTypeOrderPlaced,
				Data:           orderPlacedEventBytes,
				SequenceNumber: 0,
			},
			{
				EventType:      orders.EventTypeOrderPaid,
				Data:           orderPaidEventBytes,
				SequenceNumber: 1,
			},
		}

		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return(mockEvents, nil)
		mockProducer.On("Send", mock.Anything, mock.Anything).Return(errors.New("kafka error"))

		controller := &Controller{
			store:    mockStore,
			producer: mockProducer,
		}

		request := &pb.CancelOrderRequest{
			OrderId: "order-123",
			Reason:  "Customer requested cancellation",
		}

		response, err := controller.CancelOrder(context.Background(), request)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send order cancelled event")
		assert.Nil(t, response)
		mockStore.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})
}

func TestValidateCancelOrderRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		projection := &orders.OrderProjection{
			PaymentStatus:  orders.PaymentStatusPaid,
			ShippingStatus: orders.ShippingStatusWaitingForShipment,
		}

		err := validateCancelOrderRequest(projection)
		assert.NoError(t, err)
	})

	t.Run("order already cancelled", func(t *testing.T) {
		projection := &orders.OrderProjection{
			PaymentStatus:  orders.PaymentStatusPaid,
			ShippingStatus: orders.ShippingStatusCancelled,
		}

		err := validateCancelOrderRequest(projection)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, st.Code())
		assert.Contains(t, st.Message(), "order is already cancelled")
	})

	t.Run("order already delivered", func(t *testing.T) {
		projection := &orders.OrderProjection{
			PaymentStatus:  orders.PaymentStatusPaid,
			ShippingStatus: orders.ShippingStatusDelivered,
		}

		err := validateCancelOrderRequest(projection)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, st.Code())
		assert.Contains(t, st.Message(), "cannot cancel an order that has already been delivered")
	})
}
