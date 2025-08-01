package orders

import (
	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	orderctrl "github.com/cgund98/go-eventsrc-example/internal/entity/orders/controller"
)

type OrderService struct {
	pb.UnimplementedOrderServiceServer

	controller *orderctrl.Controller
}

func NewOrderService(controller *orderctrl.Controller) *OrderService {
	return &OrderService{controller: controller}
}
