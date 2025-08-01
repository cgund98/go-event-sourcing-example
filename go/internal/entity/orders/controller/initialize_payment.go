package controller

import (
	"context"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	"github.com/cgund98/go-eventsrc-example/internal/entity/orders"
	"github.com/cgund98/go-eventsrc-example/internal/infra/eventsrc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ErrPaymentStatusNotPending = status.Errorf(codes.FailedPrecondition, "order is not in pending payment status")

func validateInitPendingPaymentRequest(projection *orders.OrderProjection) error {

	// Check if the payment method is set
	if projection.PaymentMethod == "" {
		return status.Errorf(codes.InvalidArgument, "payment method is required")
	}

	if projection.PaymentStatus != orders.PaymentStatusPending {
		return ErrPaymentStatusNotPending
	}

	return nil
}

func (c *Controller) InitializePendingPayment(ctx context.Context, orderId string) error {

	// Fetch the order projection
	orderProjection, err := c.GetProjection(ctx, orderId)
	if err != nil {
		return err
	}
	if orderProjection == nil {
		return ErrOrderNotFound
	}

	if err := validateInitPendingPaymentRequest(orderProjection); err != nil {
		return err
	}

	// Create new event
	orderPaymentInitiatedEvent := &pb.OrderPaymentInitiated{
		OrderId:   orderId,
		Timestamp: timestamppb.Now(),
	}

	orderPaymentInitiatedEventBytes, err := proto.Marshal(orderPaymentInitiatedEvent)
	if err != nil {
		return ErrInternal
	}

	err = c.producer.Send(ctx, &eventsrc.SendArgs{
		AggregateID:   orderPaymentInitiatedEvent.OrderId,
		AggregateType: orders.AggregateTypeOrder,
		EventType:     orders.EventTypeOrderPaymentInitiated,
		Value:         orderPaymentInitiatedEventBytes,
	})
	if err != nil {
		return ErrInternal
	}

	return nil
}
