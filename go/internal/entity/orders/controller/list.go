package controller

import (
	"context"
	"fmt"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	ent "github.com/cgund98/go-eventsrc-example/internal/entity/orders"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const defaultLimit = 25
const defaultOffset = 0

func (c *Controller) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	var limit uint = defaultLimit
	var offset uint = defaultOffset
	if req.Limit != nil {
		limit = uint(*req.Limit)
	}
	if req.Offset != nil {
		offset = uint(*req.Offset)
	}

	orders, err := c.projectionRepo.List(ctx, ent.ListArgs{
		Limit:  limit,
		Offset: offset,
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
