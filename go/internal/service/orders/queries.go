package orders

import (
	"context"
	"fmt"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	"github.com/cgund98/go-eventsrc-example/internal/infra/logging"
)

func (s *OrderService) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	return WrapNonGrpcError(s.controller.ListOrders(ctx, req))
}

func (s *OrderService) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	proj, err := s.controller.GetProjection(ctx, req.OrderId)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("failed to get order projection: %v", err))
		return nil, ErrInternal()
	}

	if proj == nil {
		return &pb.GetOrderResponse{Order: nil}, nil
	}

	return &pb.GetOrderResponse{Order: proj.ToOrderDetails()}, nil
}
