package consumers

import (
	"context"
	"fmt"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	"github.com/cgund98/go-eventsrc-example/internal/entity/orders"
	"github.com/cgund98/go-eventsrc-example/internal/entity/orders/controller"
	"github.com/cgund98/go-eventsrc-example/internal/infra/logging"

	"google.golang.org/protobuf/proto"
)

const (
	ConsumerNamePaymentInitializer = "payment-initializer"
)

// PaymentInitializerConsumer is a consumer that initializes payments for orders.
// It consumes OrderPlaced events and initializes payments for those orders.
type PaymentInitializerConsumer struct {
	Controller *controller.Controller
}

func NewPaymentInitializerConsumer(controller *controller.Controller) *PaymentInitializerConsumer {
	return &PaymentInitializerConsumer{
		Controller: controller,
	}
}

func (c *PaymentInitializerConsumer) Name() string {
	return ConsumerNamePaymentInitializer
}

func (c *PaymentInitializerConsumer) Consume(ctx context.Context, eventType string, eventData []byte) error {

	// If the event type is not OrderPlaced, don't do anything
	if eventType != orders.EventTypeOrderPlaced {
		return nil
	}

	// Unmarshal the event data
	var orderPlacedEvent pb.OrderPlaced
	err := proto.Unmarshal(eventData, &orderPlacedEvent)
	if err != nil {
		return fmt.Errorf("failed to unmarshal order placed event: %w", err)
	}

	logging.Logger.Info("Initializing payment for order", "orderId", orderPlacedEvent.OrderId, "consumer", c.Name())

	err = c.Controller.InitializePendingPayment(ctx, orderPlacedEvent.OrderId)
	if err == controller.ErrPaymentStatusNotPending {
		logging.Logger.Info("Payment already initialized for order", "orderId", orderPlacedEvent.OrderId, "consumer", c.Name())
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to initialize payment: %w", err)
	}

	logging.Logger.Info("Payment initialized for order", "orderId", orderPlacedEvent.OrderId, "consumer", c.Name())

	return nil
}
