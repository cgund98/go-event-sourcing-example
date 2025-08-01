package orders

import (
	"testing"
	"time"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderProjection_ToOrderDetails(t *testing.T) {
	createdAt := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 1, 16, 14, 45, 0, 0, time.UTC)

	projection := &OrderProjection{
		OrderId:        "order-123",
		CustomerId:     "customer-456",
		VendorId:       "vendor-789",
		ProductId:      "product-101",
		Quantity:       3,
		TotalPrice:     99.99,
		PaymentMethod:  "credit_card",
		PaymentStatus:  PaymentStatusPaid,
		ShippingStatus: ShippingStatusInTransit,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}

	orderDetails := projection.ToOrderDetails()

	// Verify all fields are mapped correctly
	assert.Equal(t, "order-123", orderDetails.OrderId)
	assert.Equal(t, "customer-456", orderDetails.CustomerId)
	assert.Equal(t, "vendor-789", orderDetails.VendorId)
	assert.Equal(t, "product-101", orderDetails.ProductId)
	assert.Equal(t, int32(3), orderDetails.Quantity)
	assert.Equal(t, 99.99, orderDetails.TotalPrice)
	assert.Equal(t, "credit_card", orderDetails.PaymentMethod)
	assert.Equal(t, pb.ShippingStatus_SHIPPING_STATUS_IN_TRANSIT, orderDetails.ShippingStatus)

	// Verify timestamps
	require.NotNil(t, orderDetails.CreatedAt)
	require.NotNil(t, orderDetails.UpdatedAt)
	assert.Equal(t, createdAt, orderDetails.CreatedAt.AsTime())
	assert.Equal(t, updatedAt, orderDetails.UpdatedAt.AsTime())
}

func TestMapStrToShippingStatus(t *testing.T) {
	testCases := []struct {
		name     string
		status   string
		expected pb.ShippingStatus
	}{
		{ShippingStatusWaitingForPayment, ShippingStatusWaitingForPayment, pb.ShippingStatus_SHIPPING_STATUS_WAITING_FOR_PAYMENT},
		{ShippingStatusWaitingForShipment, ShippingStatusWaitingForShipment, pb.ShippingStatus_SHIPPING_STATUS_WAITING_FOR_SHIPMENT},
		{ShippingStatusInTransit, ShippingStatusInTransit, pb.ShippingStatus_SHIPPING_STATUS_IN_TRANSIT},
		{ShippingStatusDelivered, ShippingStatusDelivered, pb.ShippingStatus_SHIPPING_STATUS_DELIVERED},
		{ShippingStatusCancelled, ShippingStatusCancelled, pb.ShippingStatus_SHIPPING_STATUS_CANCELLED},
		{"unknown", "unknown_status", pb.ShippingStatus_SHIPPING_STATUS_UNSPECIFIED},
		{"empty", "", pb.ShippingStatus_SHIPPING_STATUS_UNSPECIFIED},
		{"uppercase", "WAITING_FOR_PAYMENT", pb.ShippingStatus_SHIPPING_STATUS_UNSPECIFIED},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapStrToShippingStatus(tc.status)
			assert.Equal(t, tc.expected, result)
		})
	}
}
