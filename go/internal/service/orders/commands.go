package orders

import (
	"context"
	"fmt"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	"github.com/cgund98/go-eventsrc-example/internal/infra/eventsrc"
	"github.com/cgund98/go-eventsrc-example/internal/infra/logging"

	"github.com/google/uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *OrderService) PlaceOrder(ctx context.Context, req *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {

	// Create new event
	orderPlacedEvent := &pb.OrderPlaced{
		OrderId:       uuid.New().String(),
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
		logging.Logger.Error(fmt.Sprintf("failed to marshal order placed event: %v", err))
		return nil, ErrInternal()
	}

	err = s.producer.Send(ctx, &eventsrc.SendArgs{
		AggregateID: orderPlacedEvent.OrderId,
		EventType:   "OrderPlaced",
		Value:       orderPlacedEventBytes,
	})
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("failed to send order placed event: %v", err))
		return nil, ErrInternal()
	}

	return &pb.PlaceOrderResponse{OrderId: orderPlacedEvent.OrderId}, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelOrder not implemented")
}

func (s *OrderService) UpdateOrderShippingStatus(ctx context.Context, req *pb.UpdateOrderShippingStatusRequest) (*pb.UpdateOrderShippingStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateOrderShippingStatus not implemented")
}
