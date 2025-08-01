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

var ErrPaymentStatusNotInitiated = status.Errorf(codes.FailedPrecondition, "order is not in initiated payment status")

func validateProcessPaymentRequest(projection *orders.OrderProjection) error {

	// Check if the payment method is set
	if projection.PaymentMethod == "" {
		return status.Errorf(codes.InvalidArgument, "payment method is required")
	}

	if projection.PaymentStatus != orders.PaymentStatusInitiated {
		return ErrPaymentStatusNotInitiated
	}

	return nil
}

func (c *Controller) ProcessPayment(ctx context.Context, orderId string) error {

	// Fetch the order projection
	orderProjection, err := c.GetProjection(ctx, orderId)
	if err != nil {
		return err
	}
	if orderProjection == nil {
		return ErrOrderNotFound
	}

	if err := validateProcessPaymentRequest(orderProjection); err != nil {
		return err
	}

	// Create new event
	orderPaymentProcessedEvent := &pb.OrderPaid{
		OrderId:   orderId,
		Timestamp: timestamppb.Now(),
	}

	orderPaymentProcessedEventBytes, err := proto.Marshal(orderPaymentProcessedEvent)
	if err != nil {
		return ErrInternal
	}

	err = c.producer.Send(ctx, &eventsrc.SendArgs{
		AggregateID:   orderPaymentProcessedEvent.OrderId,
		AggregateType: orders.AggregateTypeOrder,
		EventType:     orders.EventTypeOrderPaid,
		Value:         orderPaymentProcessedEventBytes,
	})
	if err != nil {
		return ErrInternal
	}

	return nil
}
