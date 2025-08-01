package controller

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cgund98/go-eventsrc-example/internal/entity/orders"
	"github.com/cgund98/go-eventsrc-example/internal/infra/pg"
)

// IndexProjection fetches the projection of the order and upserts it into the projection repo.
func (c *Controller) IndexProjection(ctx context.Context, orderId string) error {

	// Fetch the order projection
	orderProjection, err := c.GetProjection(ctx, orderId)
	if err != nil {
		return err
	}
	if orderProjection == nil {
		return nil
	}

	// Upsert the projection into the projection repo
	return c.transactor.WithTx(ctx, &sql.TxOptions{}, func(tx pg.Tx) error {
		err = c.projectionRepo.Upsert(ctx, tx, orders.UpsertArgs{
			OrderId:        orderId,
			PaymentStatus:  orderProjection.PaymentStatus,
			ShippingStatus: orderProjection.ShippingStatus,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert projection: %w", err)
		}
		return nil
	})
}
