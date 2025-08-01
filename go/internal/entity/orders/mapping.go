package orders

import (
	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (proj *OrderProjection) ToOrderDetails() *pb.OrderDetails {
	return &pb.OrderDetails{
		OrderId:        proj.OrderId,
		CustomerId:     proj.CustomerId,
		VendorId:       proj.VendorId,
		ProductId:      proj.ProductId,
		Quantity:       int32(proj.Quantity),
		TotalPrice:     proj.TotalPrice,
		PaymentMethod:  proj.PaymentMethod,
		ShippingStatus: MapStrToShippingStatus(proj.ShippingStatus),
		PaymentStatus:  MapStrToPaymentStatus(proj.PaymentStatus),
		CreatedAt:      timestamppb.New(proj.CreatedAt),
		UpdatedAt:      timestamppb.New(proj.UpdatedAt),
	}
}

func MapStrToShippingStatus(status string) pb.ShippingStatus {
	switch status {
	case ShippingStatusWaitingForPayment:
		return pb.ShippingStatus_SHIPPING_STATUS_WAITING_FOR_PAYMENT
	case ShippingStatusWaitingForShipment:
		return pb.ShippingStatus_SHIPPING_STATUS_WAITING_FOR_SHIPMENT
	case ShippingStatusInTransit:
		return pb.ShippingStatus_SHIPPING_STATUS_IN_TRANSIT
	case ShippingStatusDelivered:
		return pb.ShippingStatus_SHIPPING_STATUS_DELIVERED
	case ShippingStatusCancelled:
		return pb.ShippingStatus_SHIPPING_STATUS_CANCELLED
	}

	return pb.ShippingStatus_SHIPPING_STATUS_UNSPECIFIED
}

func MapShippingStatusToStr(status pb.ShippingStatus) string {
	switch status {
	case pb.ShippingStatus_SHIPPING_STATUS_WAITING_FOR_PAYMENT:
		return ShippingStatusWaitingForPayment
	case pb.ShippingStatus_SHIPPING_STATUS_WAITING_FOR_SHIPMENT:
		return ShippingStatusWaitingForShipment
	case pb.ShippingStatus_SHIPPING_STATUS_IN_TRANSIT:
		return ShippingStatusInTransit
	case pb.ShippingStatus_SHIPPING_STATUS_DELIVERED:
		return ShippingStatusDelivered
	case pb.ShippingStatus_SHIPPING_STATUS_CANCELLED:
		return ShippingStatusCancelled
	case pb.ShippingStatus_SHIPPING_STATUS_UNSPECIFIED:
		return ShippingStatusUnspecified
	}
	return ""
}

func MapStrToPaymentStatus(status string) pb.PaymentStatus {
	switch status {
	case PaymentStatusPending:
		return pb.PaymentStatus_PAYMENT_STATUS_PENDING
	case PaymentStatusInitiated:
		return pb.PaymentStatus_PAYMENT_STATUS_INITIATED
	case PaymentStatusPaid:
		return pb.PaymentStatus_PAYMENT_STATUS_PAID
	case PaymentStatusFailed:
		return pb.PaymentStatus_PAYMENT_STATUS_FAILED
	}

	return pb.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
}
