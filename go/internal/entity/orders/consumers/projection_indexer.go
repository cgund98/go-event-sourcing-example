package consumers

import (
	"context"
	"fmt"

	"github.com/cgund98/go-eventsrc-example/internal/entity/orders"
	"github.com/cgund98/go-eventsrc-example/internal/entity/orders/controller"
	"github.com/cgund98/go-eventsrc-example/internal/infra/eventsrc"
	"github.com/cgund98/go-eventsrc-example/internal/infra/logging"
)

const (
	ConsumerNameProjectionIndexer = "projection-indexer"
)

// ProjectionIndexerConsumer is a consumer that indexes projections for any order event.
type ProjectionIndexerConsumer struct {
	Controller *controller.Controller
}

func NewProjectionIndexerConsumer(controller *controller.Controller) *ProjectionIndexerConsumer {
	return &ProjectionIndexerConsumer{
		Controller: controller,
	}
}

func (c *ProjectionIndexerConsumer) Name() string {
	return ConsumerNameProjectionIndexer
}

func (c *ProjectionIndexerConsumer) Consume(ctx context.Context, args eventsrc.ConsumeArgs) error {

	if args.AggregateType != orders.AggregateTypeOrder {
		return nil
	}

	logging.Logger.Info("Indexing projection for order", "orderId", args.AggregateID, "consumer", c.Name())

	if err := c.Controller.IndexProjection(ctx, args.AggregateID); err != nil {
		return fmt.Errorf("failed to index projection: %w", err)
	}

	logging.Logger.Info("Projection indexed for order", "orderId", args.AggregateID, "consumer", c.Name())

	return nil
}
