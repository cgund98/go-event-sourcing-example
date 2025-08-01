package controller

import (
	"context"
	"fmt"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	"github.com/cgund98/go-eventsrc-example/internal/entity/orders"
	"github.com/cgund98/go-eventsrc-example/internal/infra/eventsrc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func validateUpdateShippingStatusRequest(req *pb.UpdateOrderShippingStatusRequest, projection *orders.OrderProjection) error {

	// Make sure that the order has been paid
	if projection.PaymentStatus != orders.PaymentStatusPaid {
		return status.Errorf(codes.FailedPrecondition, "order has not been paid")
	}

	// Make sure they aren't cancelling the order
	if req.Status == pb.ShippingStatus_SHIPPING_STATUS_CANCELLED {
		return status.Errorf(codes.FailedPrecondition, "cannot cancel the order when updating shipping status. Please cancel the order instead.")
	}

	// Make sure that we cannot set the shipping status to a status that is lower than the current status
	if int32(orders.MapStrToShippingStatus(projection.ShippingStatus)) > int32(req.Status) {
		return status.Errorf(codes.FailedPrecondition, "cannot set shipping status to a lower status")
	}

	return nil
}
func (c *Controller) UpdateShippingStatus(ctx context.Context, req *pb.UpdateOrderShippingStatusRequest) (*pb.UpdateOrderShippingStatusResponse, error) {

	// Fetch the order projection
	orderProjection, err := c.GetProjection(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	if orderProjection == nil {
		return nil, ErrOrderNotFound
	}

	if err := validateUpdateShippingStatusRequest(req, orderProjection); err != nil {
		return nil, err
	}

	// Create new event
	orderShippingStatusUpdatedEvent := &pb.OrderShippingStatusUpdated{
		OrderId:   req.OrderId,
		Timestamp: timestamppb.Now(),
		Status:    req.Status,
	}

	orderShippingStatusUpdatedEventBytes, err := proto.Marshal(orderShippingStatusUpdatedEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order shipping status updated event: %w", err)
	}

	err = c.producer.Send(ctx, &eventsrc.SendArgs{
		AggregateID:   req.OrderId,
		AggregateType: orders.AggregateTypeOrder,
		EventType:     orders.EventTypeOrderShippingStatusUpdated,
		Value:         orderShippingStatusUpdatedEventBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send order shipping status updated event: %w", err)
	}

	return &pb.UpdateOrderShippingStatusResponse{
		OrderId: req.OrderId,
	}, nil
}
