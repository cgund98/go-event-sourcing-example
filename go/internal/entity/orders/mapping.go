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
		ShippingStatus: mapStrToShippingStatus(proj.ShippingStatus),
		CreatedAt:      timestamppb.New(proj.CreatedAt),
		UpdatedAt:      timestamppb.New(proj.UpdatedAt),
	}
}

func mapStrToShippingStatus(status string) pb.ShippingStatus {
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
