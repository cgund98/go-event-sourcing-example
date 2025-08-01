package orders

import (
	"context"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *OrderService) PlaceOrder(ctx context.Context, req *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {
	return s.controller.PlaceOrder(ctx, req)
}

func (s *OrderService) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelOrder not implemented")
}

func (s *OrderService) UpdateOrderShippingStatus(ctx context.Context, req *pb.UpdateOrderShippingStatusRequest) (*pb.UpdateOrderShippingStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateOrderShippingStatus not implemented")
}
