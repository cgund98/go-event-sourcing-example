package controller

import (
	"context"
	"fmt"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	ent "github.com/cgund98/go-eventsrc-example/internal/entity/orders"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Controller) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	orders, err := c.projectionRepo.List(ctx, ent.ListArgs{
		Limit:  uint(req.Limit),
		Offset: uint(req.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	protoOrders := make([]*pb.ListOrdersItem, len(orders))
	for i, order := range orders {
		protoOrders[i] = &pb.ListOrdersItem{
			OrderId:        order.OrderId,
			PaymentStatus:  ent.MapStrToPaymentStatus(order.PaymentStatus),
			ShippingStatus: ent.MapStrToShippingStatus(order.ShippingStatus),
			CreatedAt:      timestamppb.New(order.CreatedAt),
			UpdatedAt:      timestamppb.New(order.UpdatedAt),
		}
	}

	return &pb.ListOrdersResponse{
		Orders: protoOrders,
	}, nil
}
