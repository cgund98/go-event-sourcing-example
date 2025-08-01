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

func createValidOrderPaidEvent(orderId string) []byte {
	event := &pb.OrderPaid{
		OrderId:   orderId,
		Timestamp: timestamppb.Now(),
	}

	eventBytes, _ := proto.Marshal(event)
	return eventBytes
}

func TestController_UpdateShippingStatus(t *testing.T) {
	t.Run("successful shipping status update", func(t *testing.T) {
		mockStore := &MockStore{}
		mockProducer := &MockProducer{}

		// Create events: OrderPlaced -> OrderPaid (creates paid status)
		orderPlacedEventBytes := createValidOrderPlacedEvent("order-123", "credit_card")
		orderPaidEventBytes := createValidOrderPaidEvent("order-123")

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
		mockProducer.On("Send", mock.Anything, mock.MatchedBy(func(args *eventsrc.SendArgs) bool {
			return args.AggregateID == "order-123" &&
				args.AggregateType == orders.AggregateTypeOrder &&
				args.EventType == orders.EventTypeOrderShippingStatusUpdated &&
				len(args.Value) > 0
		})).Return(nil)

		controller := &Controller{
			store:    mockStore,
			producer: mockProducer,
		}

		request := &pb.UpdateOrderShippingStatusRequest{
			OrderId: "order-123",
			Status:  pb.ShippingStatus_SHIPPING_STATUS_IN_TRANSIT,
		}

		response, err := controller.UpdateShippingStatus(context.Background(), request)

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

		request := &pb.UpdateOrderShippingStatusRequest{
			OrderId: "order-123",
			Status:  pb.ShippingStatus_SHIPPING_STATUS_IN_TRANSIT,
		}

		response, err := controller.UpdateShippingStatus(context.Background(), request)

		assert.Error(t, err)
		assert.Equal(t, ErrOrderNotFound, err)
		assert.Nil(t, response)
		mockStore.AssertExpectations(t)
	})

	t.Run("order not paid", func(t *testing.T) {
		mockStore := &MockStore{}

		// Only OrderPlaced event (creates pending payment status)
		orderPlacedEventBytes := createValidOrderPlacedEvent("order-123", "credit_card")
		mockEvents := []eventsrc.Event{
			{
				EventType: orders.EventTypeOrderPlaced,
				Data:      orderPlacedEventBytes,
			},
		}

		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return(mockEvents, nil)

		controller := &Controller{store: mockStore}

		request := &pb.UpdateOrderShippingStatusRequest{
			OrderId: "order-123",
			Status:  pb.ShippingStatus_SHIPPING_STATUS_IN_TRANSIT,
		}

		response, err := controller.UpdateShippingStatus(context.Background(), request)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, st.Code())
		assert.Contains(t, st.Message(), "order has not been paid")
		assert.Nil(t, response)
		mockStore.AssertExpectations(t)
	})

	t.Run("lower shipping status", func(t *testing.T) {
		mockStore := &MockStore{}

		// Create events: OrderPlaced -> OrderPaid -> ShippingStatusUpdated (in_transit)
		orderPlacedEventBytes := createValidOrderPlacedEvent("order-123", "credit_card")
		orderPaidEventBytes := createValidOrderPaidEvent("order-123")

		shippingEvent := &pb.OrderShippingStatusUpdated{
			OrderId:   "order-123",
			Timestamp: timestamppb.Now(),
			Status:    pb.ShippingStatus_SHIPPING_STATUS_IN_TRANSIT,
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

		request := &pb.UpdateOrderShippingStatusRequest{
			OrderId: "order-123",
			Status:  pb.ShippingStatus_SHIPPING_STATUS_WAITING_FOR_SHIPMENT, // Lower than in_transit
		}

		response, err := controller.UpdateShippingStatus(context.Background(), request)

		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, st.Code())
		assert.Contains(t, st.Message(), "cannot set shipping status to a lower status")
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
				EventType: orders.EventTypeOrderPlaced,
				Data:      orderPlacedEventBytes,
			},
			{
				EventType: orders.EventTypeOrderPaid,
				Data:      orderPaidEventBytes,
			},
		}

		mockStore.On("ListByAggregateID", mock.Anything, "order-123", orders.AggregateTypeOrder).Return(mockEvents, nil)
		mockProducer.On("Send", mock.Anything, mock.Anything).Return(errors.New("kafka error"))

		controller := &Controller{
			store:    mockStore,
			producer: mockProducer,
		}

		request := &pb.UpdateOrderShippingStatusRequest{
			OrderId: "order-123",
			Status:  pb.ShippingStatus_SHIPPING_STATUS_IN_TRANSIT,
		}

		response, err := controller.UpdateShippingStatus(context.Background(), request)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send order shipping status updated event")
		assert.Nil(t, response)
		mockStore.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})
}

func TestValidateUpdateShippingStatusRequest(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		projection := &orders.OrderProjection{
			PaymentStatus:  orders.PaymentStatusPaid,
			ShippingStatus: orders.ShippingStatusWaitingForShipment,
		}

		request := &pb.UpdateOrderShippingStatusRequest{
			OrderId: "order-123",
			Status:  pb.ShippingStatus_SHIPPING_STATUS_IN_TRANSIT,
		}

		err := validateUpdateShippingStatusRequest(request, projection)
		assert.NoError(t, err)
	})

	t.Run("order not paid", func(t *testing.T) {
		projection := &orders.OrderProjection{
			PaymentStatus:  orders.PaymentStatusPending,
			ShippingStatus: orders.ShippingStatusWaitingForPayment,
		}

		request := &pb.UpdateOrderShippingStatusRequest{
			OrderId: "order-123",
			Status:  pb.ShippingStatus_SHIPPING_STATUS_IN_TRANSIT,
		}

		err := validateUpdateShippingStatusRequest(request, projection)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, st.Code())
		assert.Contains(t, st.Message(), "order has not been paid")
	})

	t.Run("lower shipping status", func(t *testing.T) {
		projection := &orders.OrderProjection{
			PaymentStatus:  orders.PaymentStatusPaid,
			ShippingStatus: orders.ShippingStatusInTransit,
		}

		request := &pb.UpdateOrderShippingStatusRequest{
			OrderId: "order-123",
			Status:  pb.ShippingStatus_SHIPPING_STATUS_WAITING_FOR_SHIPMENT, // Lower than in_transit
		}

		err := validateUpdateShippingStatusRequest(request, projection)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.FailedPrecondition, st.Code())
		assert.Contains(t, st.Message(), "cannot set shipping status to a lower status")
	})
}
