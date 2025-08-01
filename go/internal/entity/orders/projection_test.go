package orders

import (
	"testing"
	"time"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestApplyOrderPlacedToProjection(t *testing.T) {
	// Create test event data
	timestamp := time.Now().UTC()
	event := &pb.OrderPlaced{
		OrderId:       "order-123",
		CustomerId:    "customer-456",
		VendorId:      "vendor-789",
		ProductId:     "product-101",
		Quantity:      5,
		TotalPrice:    99.99,
		PaymentMethod: "credit_card",
		Timestamp:     timestamppb.New(timestamp),
	}

	eventData, err := proto.Marshal(event)
	require.NoError(t, err)

	// Test projection
	projection := &OrderProjection{}
	err = applyOrderPlacedToProjection(eventData, projection)
	require.NoError(t, err)

	// Verify all fields are set correctly
	assert.Equal(t, "order-123", projection.OrderId)
	assert.Equal(t, "customer-456", projection.CustomerId)
	assert.Equal(t, "vendor-789", projection.VendorId)
	assert.Equal(t, "product-101", projection.ProductId)
	assert.Equal(t, int32(5), projection.Quantity)
	assert.Equal(t, 99.99, projection.TotalPrice)
	assert.Equal(t, "credit_card", projection.PaymentMethod)
	assert.Equal(t, PaymentStatusPending, projection.PaymentStatus)
	assert.Equal(t, ShippingStatusWaitingForPayment, projection.ShippingStatus)
	assert.Equal(t, timestamp, projection.CreatedAt)
	assert.Equal(t, timestamp, projection.UpdatedAt)
}

func TestApplyOrderPaymentInitiatedToProjection(t *testing.T) {
	// Create test event data
	timestamp := time.Now().UTC()
	event := &pb.OrderPaymentInitiated{
		OrderId:   "order-123",
		Timestamp: timestamppb.New(timestamp),
	}

	eventData, err := proto.Marshal(event)
	require.NoError(t, err)

	// Test projection
	originalCreatedAt := time.Now().Add(-1 * time.Hour).UTC()
	projection := &OrderProjection{
		OrderId:        "order-123",
		PaymentStatus:  PaymentStatusPending,
		ShippingStatus: ShippingStatusWaitingForPayment,
		CreatedAt:      originalCreatedAt,
		UpdatedAt:      originalCreatedAt,
	}

	err = applyOrderPaymentInitiatedToProjection(eventData, projection)
	require.NoError(t, err)

	// Verify payment status is updated
	assert.Equal(t, PaymentStatusInitiated, projection.PaymentStatus)
	// Verify UpdatedAt is updated
	assert.Equal(t, timestamp, projection.UpdatedAt)
	// Verify other fields remain unchanged
	assert.Equal(t, "order-123", projection.OrderId)
	assert.Equal(t, ShippingStatusWaitingForPayment, projection.ShippingStatus)
	assert.Equal(t, originalCreatedAt, projection.CreatedAt)
}

func TestApplyOrderPaidToProjection(t *testing.T) {
	// Create test event data
	timestamp := time.Now().UTC()
	event := &pb.OrderPaid{
		OrderId:   "order-123",
		Timestamp: timestamppb.New(timestamp),
	}

	eventData, err := proto.Marshal(event)
	require.NoError(t, err)

	// Test projection
	originalCreatedAt := time.Now().Add(-1 * time.Hour).UTC()
	projection := &OrderProjection{
		OrderId:        "order-123",
		PaymentStatus:  PaymentStatusInitiated,
		ShippingStatus: ShippingStatusWaitingForPayment,
		CreatedAt:      originalCreatedAt,
		UpdatedAt:      originalCreatedAt,
	}

	err = applyOrderPaidToProjection(eventData, projection)
	require.NoError(t, err)

	// Verify both payment and shipping status are updated
	assert.Equal(t, PaymentStatusPaid, projection.PaymentStatus)
	assert.Equal(t, ShippingStatusWaitingForShipment, projection.ShippingStatus)
	// Verify UpdatedAt is updated
	assert.Equal(t, timestamp, projection.UpdatedAt)
	// Verify other fields remain unchanged
	assert.Equal(t, "order-123", projection.OrderId)
	assert.Equal(t, originalCreatedAt, projection.CreatedAt)
}

func TestApplyOrderPaymentFailedToProjection(t *testing.T) {
	// Create test event data
	timestamp := time.Now().UTC()
	event := &pb.OrderPaymentFailed{
		OrderId:   "order-123",
		Timestamp: timestamppb.New(timestamp),
	}

	eventData, err := proto.Marshal(event)
	require.NoError(t, err)

	// Test projection
	originalCreatedAt := time.Now().Add(-1 * time.Hour).UTC()
	projection := &OrderProjection{
		OrderId:        "order-123",
		PaymentStatus:  PaymentStatusInitiated,
		ShippingStatus: ShippingStatusWaitingForPayment,
		CreatedAt:      originalCreatedAt,
		UpdatedAt:      originalCreatedAt,
	}

	err = applyOrderPaymentFailedToProjection(eventData, projection)
	require.NoError(t, err)

	// Verify payment status is updated to failed
	assert.Equal(t, PaymentStatusFailed, projection.PaymentStatus)
	// Verify UpdatedAt is updated
	assert.Equal(t, timestamp, projection.UpdatedAt)
	// Verify other fields remain unchanged
	assert.Equal(t, "order-123", projection.OrderId)
	assert.Equal(t, ShippingStatusWaitingForPayment, projection.ShippingStatus)
	assert.Equal(t, originalCreatedAt, projection.CreatedAt)
}

func TestApplyOrderCancelledToProjection(t *testing.T) {
	// Create test event data
	timestamp := time.Now().UTC()
	event := &pb.OrderCancelled{
		OrderId:   "order-123",
		Timestamp: timestamppb.New(timestamp),
	}

	eventData, err := proto.Marshal(event)
	require.NoError(t, err)

	// Test projection
	originalCreatedAt := time.Now().Add(-1 * time.Hour).UTC()
	projection := &OrderProjection{
		OrderId:        "order-123",
		PaymentStatus:  PaymentStatusPaid,
		ShippingStatus: ShippingStatusWaitingForShipment,
		CreatedAt:      originalCreatedAt,
		UpdatedAt:      originalCreatedAt,
	}

	err = applyOrderCancelledToProjection(eventData, projection)
	require.NoError(t, err)

	// Verify shipping status is updated to cancelled
	assert.Equal(t, ShippingStatusCancelled, projection.ShippingStatus)
	// Verify UpdatedAt is updated
	assert.Equal(t, timestamp, projection.UpdatedAt)
	// Verify other fields remain unchanged
	assert.Equal(t, "order-123", projection.OrderId)
	assert.Equal(t, PaymentStatusPaid, projection.PaymentStatus)
	assert.Equal(t, originalCreatedAt, projection.CreatedAt)
}

func TestApplyOrderShippingStatusUpdatedToProjection_Delivered(t *testing.T) {
	// Create test event data
	timestamp := time.Now().UTC()
	event := &pb.OrderShippingStatusUpdated{
		OrderId:   "order-123",
		Status:    pb.ShippingStatus_SHIPPING_STATUS_DELIVERED,
		Timestamp: timestamppb.New(timestamp),
	}

	eventData, err := proto.Marshal(event)
	require.NoError(t, err)

	// Test projection
	originalCreatedAt := time.Now().Add(-1 * time.Hour).UTC()
	projection := &OrderProjection{
		OrderId:        "order-123",
		PaymentStatus:  PaymentStatusPaid,
		ShippingStatus: ShippingStatusInTransit,
		CreatedAt:      originalCreatedAt,
		UpdatedAt:      originalCreatedAt,
	}

	err = applyOrderShippingStatusUpdatedToProjection(eventData, projection)
	require.NoError(t, err)

	// Verify shipping status is updated to delivered
	assert.Equal(t, ShippingStatusDelivered, projection.ShippingStatus)
	// Verify UpdatedAt is updated
	assert.Equal(t, timestamp, projection.UpdatedAt)
	// Verify other fields remain unchanged
	assert.Equal(t, "order-123", projection.OrderId)
	assert.Equal(t, PaymentStatusPaid, projection.PaymentStatus)
	assert.Equal(t, originalCreatedAt, projection.CreatedAt)
}

func TestApplyOrderShippingStatusUpdatedToProjection_InTransit(t *testing.T) {
	// Create test event data
	timestamp := time.Now().UTC()
	event := &pb.OrderShippingStatusUpdated{
		OrderId:   "order-123",
		Status:    pb.ShippingStatus_SHIPPING_STATUS_IN_TRANSIT,
		Timestamp: timestamppb.New(timestamp),
	}

	eventData, err := proto.Marshal(event)
	require.NoError(t, err)

	// Test projection
	originalCreatedAt := time.Now().Add(-1 * time.Hour).UTC()
	projection := &OrderProjection{
		OrderId:        "order-123",
		PaymentStatus:  PaymentStatusPaid,
		ShippingStatus: ShippingStatusWaitingForShipment,
		CreatedAt:      originalCreatedAt,
		UpdatedAt:      originalCreatedAt,
	}

	err = applyOrderShippingStatusUpdatedToProjection(eventData, projection)
	require.NoError(t, err)

	// Verify shipping status is updated to in transit
	assert.Equal(t, ShippingStatusInTransit, projection.ShippingStatus)
	// Verify UpdatedAt is updated
	assert.Equal(t, timestamp, projection.UpdatedAt)
	// Verify other fields remain unchanged
	assert.Equal(t, "order-123", projection.OrderId)
	assert.Equal(t, PaymentStatusPaid, projection.PaymentStatus)
	assert.Equal(t, originalCreatedAt, projection.CreatedAt)
}

func TestApplyOrderShippingStatusUpdatedToProjection_UnknownStatus(t *testing.T) {
	// Create test event data with unknown status
	timestamp := time.Now().UTC()
	event := &pb.OrderShippingStatusUpdated{
		OrderId:   "order-123",
		Status:    pb.ShippingStatus_SHIPPING_STATUS_UNSPECIFIED,
		Timestamp: timestamppb.New(timestamp),
	}

	eventData, err := proto.Marshal(event)
	require.NoError(t, err)

	// Test projection
	originalCreatedAt := time.Now().Add(-1 * time.Hour).UTC()
	projection := &OrderProjection{
		OrderId:        "order-123",
		PaymentStatus:  PaymentStatusPaid,
		ShippingStatus: ShippingStatusWaitingForShipment,
		CreatedAt:      originalCreatedAt,
		UpdatedAt:      originalCreatedAt,
	}

	err = applyOrderShippingStatusUpdatedToProjection(eventData, projection)
	require.NoError(t, err)

	// Verify shipping status remains unchanged for unknown status
	assert.Equal(t, ShippingStatusWaitingForShipment, projection.ShippingStatus)
	// Verify UpdatedAt is still updated even for unknown status
	assert.Equal(t, timestamp, projection.UpdatedAt)
	// Verify other fields remain unchanged
	assert.Equal(t, "order-123", projection.OrderId)
	assert.Equal(t, PaymentStatusPaid, projection.PaymentStatus)
	assert.Equal(t, originalCreatedAt, projection.CreatedAt)
}

func TestReduceToProjection_SingleEvent(t *testing.T) {
	// Create a single OrderPlaced event
	timestamp := time.Now().UTC()
	orderPlacedEvent := &pb.OrderPlaced{
		OrderId:       "order-123",
		CustomerId:    "customer-456",
		VendorId:      "vendor-789",
		ProductId:     "product-101",
		Quantity:      3,
		TotalPrice:    150.00,
		PaymentMethod: "paypal",
		Timestamp:     timestamppb.New(timestamp),
	}

	eventData, err := proto.Marshal(orderPlacedEvent)
	require.NoError(t, err)

	events := []SerializedEvent{
		{
			EventType: EventTypeOrderPlaced,
			EventData: eventData,
		},
	}

	// Test reduction
	projection, err := ReduceToProjection(events)
	require.NoError(t, err)

	// Verify projection
	assert.Equal(t, "order-123", projection.OrderId)
	assert.Equal(t, "customer-456", projection.CustomerId)
	assert.Equal(t, "vendor-789", projection.VendorId)
	assert.Equal(t, "product-101", projection.ProductId)
	assert.Equal(t, int32(3), projection.Quantity)
	assert.Equal(t, 150.00, projection.TotalPrice)
	assert.Equal(t, "paypal", projection.PaymentMethod)
	assert.Equal(t, PaymentStatusPending, projection.PaymentStatus)
	assert.Equal(t, ShippingStatusWaitingForPayment, projection.ShippingStatus)
	assert.Equal(t, timestamp, projection.CreatedAt)
	assert.Equal(t, timestamp, projection.UpdatedAt)
}

func TestReduceToProjection_MultipleEvents(t *testing.T) {
	// Create multiple events in sequence with different timestamps
	baseTime := time.Now().UTC()

	orderPlacedEvent := &pb.OrderPlaced{
		OrderId:       "order-123",
		CustomerId:    "customer-456",
		VendorId:      "vendor-789",
		ProductId:     "product-101",
		Quantity:      2,
		TotalPrice:    100.00,
		PaymentMethod: "credit_card",
		Timestamp:     timestamppb.New(baseTime),
	}

	paymentInitiatedEvent := &pb.OrderPaymentInitiated{
		OrderId:   "order-123",
		Timestamp: timestamppb.New(baseTime.Add(1 * time.Hour)),
	}

	orderPaidEvent := &pb.OrderPaid{
		OrderId:   "order-123",
		Timestamp: timestamppb.New(baseTime.Add(2 * time.Hour)),
	}

	shippingUpdatedEvent := &pb.OrderShippingStatusUpdated{
		OrderId:   "order-123",
		Status:    pb.ShippingStatus_SHIPPING_STATUS_IN_TRANSIT,
		Timestamp: timestamppb.New(baseTime.Add(3 * time.Hour)),
	}

	// Marshal all events
	orderPlacedData, err := proto.Marshal(orderPlacedEvent)
	require.NoError(t, err)
	paymentInitiatedData, err := proto.Marshal(paymentInitiatedEvent)
	require.NoError(t, err)
	orderPaidData, err := proto.Marshal(orderPaidEvent)
	require.NoError(t, err)
	shippingUpdatedData, err := proto.Marshal(shippingUpdatedEvent)
	require.NoError(t, err)

	events := []SerializedEvent{
		{
			EventType: EventTypeOrderPlaced,
			EventData: orderPlacedData,
		},
		{
			EventType: EventTypeOrderPaymentInitiated,
			EventData: paymentInitiatedData,
		},
		{
			EventType: EventTypeOrderPaid,
			EventData: orderPaidData,
		},
		{
			EventType: EventTypeOrderShippingStatusUpdated,
			EventData: shippingUpdatedData,
		},
	}

	// Test reduction
	projection, err := ReduceToProjection(events)
	require.NoError(t, err)

	// Verify final projection state
	assert.Equal(t, "order-123", projection.OrderId)
	assert.Equal(t, "customer-456", projection.CustomerId)
	assert.Equal(t, "vendor-789", projection.VendorId)
	assert.Equal(t, "product-101", projection.ProductId)
	assert.Equal(t, int32(2), projection.Quantity)
	assert.Equal(t, 100.00, projection.TotalPrice)
	assert.Equal(t, "credit_card", projection.PaymentMethod)
	assert.Equal(t, PaymentStatusPaid, projection.PaymentStatus)
	assert.Equal(t, ShippingStatusInTransit, projection.ShippingStatus)
	assert.Equal(t, baseTime, projection.CreatedAt)
	assert.Equal(t, baseTime.Add(3*time.Hour), projection.UpdatedAt)
}

func TestReduceToProjection_EmptyEvents(t *testing.T) {
	events := []SerializedEvent{}

	projection, err := ReduceToProjection(events)
	require.NoError(t, err)

	// Verify empty projection
	assert.Equal(t, "", projection.OrderId)
	assert.Equal(t, "", projection.CustomerId)
	assert.Equal(t, "", projection.VendorId)
	assert.Equal(t, "", projection.ProductId)
	assert.Equal(t, int32(0), projection.Quantity)
	assert.Equal(t, 0.0, projection.TotalPrice)
	assert.Equal(t, "", projection.PaymentMethod)
	assert.Equal(t, "", projection.PaymentStatus)
	assert.Equal(t, "", projection.ShippingStatus)
	assert.Equal(t, time.Time{}, projection.CreatedAt)
	assert.Equal(t, time.Time{}, projection.UpdatedAt)
}

func TestReduceToProjection_UnknownEventType(t *testing.T) {
	events := []SerializedEvent{
		{
			EventType: "UnknownEventType",
			EventData: []byte{},
		},
	}

	_, err := ReduceToProjection(events)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown event type")
}

func TestReduceToProjection_InvalidEventData(t *testing.T) {
	events := []SerializedEvent{
		{
			EventType: EventTypeOrderPlaced,
			EventData: []byte("invalid protobuf data"),
		},
	}

	_, err := ReduceToProjection(events)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")
}

func TestReduceToProjection_CancelledOrder(t *testing.T) {
	// Create events: OrderPlaced -> OrderPaid -> OrderCancelled
	baseTime := time.Now().UTC()

	orderPlacedEvent := &pb.OrderPlaced{
		OrderId:       "order-123",
		CustomerId:    "customer-456",
		VendorId:      "vendor-789",
		ProductId:     "product-101",
		Quantity:      1,
		TotalPrice:    50.00,
		PaymentMethod: "debit_card",
		Timestamp:     timestamppb.New(baseTime),
	}

	orderPaidEvent := &pb.OrderPaid{
		OrderId:   "order-123",
		Timestamp: timestamppb.New(baseTime.Add(1 * time.Hour)),
	}

	orderCancelledEvent := &pb.OrderCancelled{
		OrderId:   "order-123",
		Timestamp: timestamppb.New(baseTime.Add(2 * time.Hour)),
	}

	// Marshal events
	orderPlacedData, err := proto.Marshal(orderPlacedEvent)
	require.NoError(t, err)
	orderPaidData, err := proto.Marshal(orderPaidEvent)
	require.NoError(t, err)
	orderCancelledData, err := proto.Marshal(orderCancelledEvent)
	require.NoError(t, err)

	events := []SerializedEvent{
		{
			EventType: EventTypeOrderPlaced,
			EventData: orderPlacedData,
		},
		{
			EventType: EventTypeOrderPaid,
			EventData: orderPaidData,
		},
		{
			EventType: EventTypeOrderCancelled,
			EventData: orderCancelledData,
		},
	}

	// Test reduction
	projection, err := ReduceToProjection(events)
	require.NoError(t, err)

	// Verify final state shows cancelled shipping status
	assert.Equal(t, "order-123", projection.OrderId)
	assert.Equal(t, PaymentStatusPaid, projection.PaymentStatus)
	assert.Equal(t, ShippingStatusCancelled, projection.ShippingStatus)
	assert.Equal(t, baseTime, projection.CreatedAt)
	assert.Equal(t, baseTime.Add(2*time.Hour), projection.UpdatedAt)
}
