package consumers

import (
	"context"
	"fmt"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	"github.com/cgund98/go-eventsrc-example/internal/entity/orders"
	"github.com/cgund98/go-eventsrc-example/internal/entity/orders/controller"
	"github.com/cgund98/go-eventsrc-example/internal/infra/eventsrc"
	"github.com/cgund98/go-eventsrc-example/internal/infra/logging"

	"google.golang.org/protobuf/proto"
)

const (
	ConsumerNamePaymentProcessor = "payment-processor"
)

// PaymentProcessorConsumer is a consumer that initializes payments for orders.
// It consumes OrderPaymentInitiated events and processes payments for those orders.
type PaymentProcessorConsumer struct {
	Controller *controller.Controller
}

func NewPaymentProcessorConsumer(controller *controller.Controller) *PaymentProcessorConsumer {
	return &PaymentProcessorConsumer{
		Controller: controller,
	}
}

func (c *PaymentProcessorConsumer) Name() string {
	return ConsumerNamePaymentProcessor
}

func (c *PaymentProcessorConsumer) Consume(ctx context.Context, args eventsrc.ConsumeArgs) error {

	if args.AggregateType != orders.AggregateTypeOrder {
		return nil
	}

	// If the event type is not OrderPaymentInitiated, don't do anything
	if args.EventType != orders.EventTypeOrderPaymentInitiated {
		return nil
	}

	// Unmarshal the event data
	var orderPlacedEvent pb.OrderPaymentInitiated
	err := proto.Unmarshal(args.Data, &orderPlacedEvent)
	if err != nil {
		return fmt.Errorf("failed to unmarshal order placed event: %w", err)
	}

	logging.Logger.Info("Processing payment for order", "orderId", orderPlacedEvent.OrderId, "consumer", c.Name())

	err = c.Controller.ProcessPayment(ctx, orderPlacedEvent.OrderId)
	if err == controller.ErrPaymentStatusNotInitiated {
		logging.Logger.Info("Payment already processed for order", "orderId", orderPlacedEvent.OrderId, "consumer", c.Name())
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to process payment: %w", err)
	}

	logging.Logger.Info("Payment processed for order", "orderId", orderPlacedEvent.OrderId, "consumer", c.Name())

	return nil
}
