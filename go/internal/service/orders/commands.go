package orders

import (
	"context"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
)

func (s *OrderService) PlaceOrder(ctx context.Context, req *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {
	return WrapNonGrpcError(s.controller.PlaceOrder(ctx, req))
}

func (s *OrderService) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	return WrapNonGrpcError(s.controller.CancelOrder(ctx, req))
}

func (s *OrderService) UpdateOrderShippingStatus(ctx context.Context, req *pb.UpdateOrderShippingStatusRequest) (*pb.UpdateOrderShippingStatusResponse, error) {
	return WrapNonGrpcError(s.controller.UpdateShippingStatus(ctx, req))
}
