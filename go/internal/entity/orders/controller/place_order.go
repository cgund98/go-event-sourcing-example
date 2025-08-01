package controller

import (
	"context"
	"fmt"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	"github.com/cgund98/go-eventsrc-example/internal/entity/orders"
	"github.com/cgund98/go-eventsrc-example/internal/infra/eventsrc"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Controller) PlaceOrder(ctx context.Context, req *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {
	orderId, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate order id: %w", err)
	}

	// Create new event
	orderPlacedEvent := &pb.OrderPlaced{
		OrderId:       orderId.String(),
		Timestamp:     timestamppb.Now(),
		VendorId:      req.VendorId,
		CustomerId:    req.CustomerId,
		ProductId:     req.ProductId,
		Quantity:      req.Quantity,
		TotalPrice:    req.TotalPrice,
		PaymentMethod: req.PaymentMethod,
	}

	orderPlacedEventBytes, err := proto.Marshal(orderPlacedEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order placed event: %w", err)
	}

	err = c.producer.Send(ctx, &eventsrc.SendArgs{
		AggregateID:   orderPlacedEvent.OrderId,
		AggregateType: orders.AggregateTypeOrder,
		EventType:     orders.EventTypeOrderPlaced,
		Value:         orderPlacedEventBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send order placed event: %w", err)
	}

	return &pb.PlaceOrderResponse{OrderId: orderPlacedEvent.OrderId}, nil
}
