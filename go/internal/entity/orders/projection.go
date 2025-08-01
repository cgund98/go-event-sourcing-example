package orders

import (
	"fmt"
	"time"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	"google.golang.org/protobuf/proto"
)

const (
	// Payment status enum
	PaymentStatusPending   = "pending"
	PaymentStatusInitiated = "initiated"
	PaymentStatusPaid      = "paid"
	PaymentStatusFailed    = "failed"

	// Shipping status enum
	ShippingStatusUnspecified        = "unspecified"
	ShippingStatusWaitingForPayment  = "waiting_for_payment"
	ShippingStatusWaitingForShipment = "waiting_for_shipment"
	ShippingStatusInTransit          = "in_transit"
	ShippingStatusDelivered          = "delivered"
	ShippingStatusCancelled          = "cancelled"
)

type OrderProjection struct {
	OrderId        string
	CustomerId     string
	VendorId       string
	ProductId      string
	Quantity       int32
	TotalPrice     float64
	PaymentMethod  string
	PaymentStatus  string
	ShippingStatus string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type SerializedEvent struct {
	EventType string
	EventData []byte
}

// ReduceToProjection will reduce the event list into an order projection
func ReduceToProjection(events []SerializedEvent) (*OrderProjection, error) {
	projection := &OrderProjection{}
	for _, event := range events {
		err := applyEventToProjection(event, projection)
		if err != nil {
			return nil, err
		}
	}
	return projection, nil
}

// applyEventToProjection will map the event type to the appropriate apply function
func applyEventToProjection(event SerializedEvent, currentProjection *OrderProjection) error {
	switch event.EventType {
	case EventTypeOrderPlaced:
		return applyOrderPlacedToProjection(event.EventData, currentProjection)
	case EventTypeOrderPaymentInitiated:
		return applyOrderPaymentInitiatedToProjection(event.EventData, currentProjection)
	case EventTypeOrderPaid:
		return applyOrderPaidToProjection(event.EventData, currentProjection)
	case EventTypeOrderPaymentFailed:
		return applyOrderPaymentFailedToProjection(event.EventData, currentProjection)
	case EventTypeOrderCancelled:
		return applyOrderCancelledToProjection(event.EventData, currentProjection)
	case EventTypeOrderShippingStatusUpdated:
		return applyOrderShippingStatusUpdatedToProjection(event.EventData, currentProjection)
	default:
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}
}

func applyOrderPlacedToProjection(eventData []byte, currentProjection *OrderProjection) error {
	var event pb.OrderPlaced
	err := proto.Unmarshal(eventData, &event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal order placed event: %w", err)
	}

	currentProjection.OrderId = event.OrderId
	currentProjection.CustomerId = event.CustomerId
	currentProjection.VendorId = event.VendorId
	currentProjection.ProductId = event.ProductId
	currentProjection.Quantity = event.Quantity
	currentProjection.TotalPrice = event.TotalPrice
	currentProjection.PaymentMethod = event.PaymentMethod
	currentProjection.PaymentStatus = PaymentStatusPending
	currentProjection.ShippingStatus = ShippingStatusWaitingForPayment
	currentProjection.CreatedAt = event.Timestamp.AsTime()
	currentProjection.UpdatedAt = event.Timestamp.AsTime()

	return nil
}

func applyOrderPaymentInitiatedToProjection(eventData []byte, currentProjection *OrderProjection) error {
	var event pb.OrderPaymentInitiated
	err := proto.Unmarshal(eventData, &event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal order payment initiated event: %w", err)
	}

	currentProjection.PaymentStatus = PaymentStatusInitiated
	currentProjection.UpdatedAt = event.Timestamp.AsTime()

	return nil
}

func applyOrderPaidToProjection(eventData []byte, currentProjection *OrderProjection) error {
	var event pb.OrderPaid
	err := proto.Unmarshal(eventData, &event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal order paid event: %w", err)
	}

	currentProjection.PaymentStatus = PaymentStatusPaid
	currentProjection.ShippingStatus = ShippingStatusWaitingForShipment
	currentProjection.UpdatedAt = event.Timestamp.AsTime()

	return nil
}

func applyOrderPaymentFailedToProjection(eventData []byte, currentProjection *OrderProjection) error {
	var event pb.OrderPaymentFailed
	err := proto.Unmarshal(eventData, &event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal order payment failed event: %w", err)
	}

	currentProjection.PaymentStatus = PaymentStatusFailed
	currentProjection.UpdatedAt = event.Timestamp.AsTime()
	return nil
}

func applyOrderCancelledToProjection(eventData []byte, currentProjection *OrderProjection) error {
	var event pb.OrderCancelled
	err := proto.Unmarshal(eventData, &event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal order cancelled event: %w", err)
	}

	currentProjection.ShippingStatus = ShippingStatusCancelled
	currentProjection.UpdatedAt = event.Timestamp.AsTime()

	return nil
}

func applyOrderShippingStatusUpdatedToProjection(eventData []byte, currentProjection *OrderProjection) error {
	var event pb.OrderShippingStatusUpdated
	err := proto.Unmarshal(eventData, &event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal order shipping status updated event: %w", err)
	}

	if event.Status != pb.ShippingStatus_SHIPPING_STATUS_UNSPECIFIED {
		currentProjection.ShippingStatus = MapShippingStatusToStr(event.Status)
	}

	currentProjection.UpdatedAt = event.Timestamp.AsTime()

	return nil
}
