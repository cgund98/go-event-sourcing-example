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

func validateCancelOrderRequest(projection *orders.OrderProjection) error {
	// Only allow cancelling if the order is paid and not already cancelled or shipped
	if projection.ShippingStatus == orders.ShippingStatusCancelled {
		return status.Errorf(codes.FailedPrecondition, "order is already cancelled")
	}
	if projection.ShippingStatus == orders.ShippingStatusDelivered {
		return status.Errorf(codes.FailedPrecondition, "cannot cancel an order that has already been delivered")
	}
	return nil
}

func (c *Controller) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	// Fetch the order projection
	orderProjection, err := c.GetProjection(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}
	if orderProjection == nil {
		return nil, ErrOrderNotFound
	}

	if err := validateCancelOrderRequest(orderProjection); err != nil {
		return nil, err
	}

	// Create new event
	orderCancelledEvent := &pb.OrderCancelled{
		OrderId:   req.OrderId,
		Timestamp: timestamppb.Now(),
		Reason:    req.Reason,
	}

	orderCancelledEventBytes, err := proto.Marshal(orderCancelledEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order cancelled event: %w", err)
	}

	err = c.producer.Send(ctx, &eventsrc.SendArgs{
		AggregateID:   req.OrderId,
		AggregateType: orders.AggregateTypeOrder,
		EventType:     orders.EventTypeOrderCancelled,
		Value:         orderCancelledEventBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send order cancelled event: %w", err)
	}

	return &pb.CancelOrderResponse{
		OrderId: req.OrderId,
	}, nil
}
