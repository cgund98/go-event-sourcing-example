package orders

import (
	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"

	"github.com/cgund98/go-eventsrc-example/internal/infra/eventsrc"
)

type OrderService struct {
	pb.UnimplementedOrderServiceServer

	producer *eventsrc.TransactionProducer
}

func NewOrderService(producer *eventsrc.TransactionProducer) *OrderService {
	return &OrderService{producer: producer}
}
