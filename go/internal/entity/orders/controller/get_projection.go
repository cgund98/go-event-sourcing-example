package controller

import (
	"context"
	"fmt"

	"github.com/cgund98/go-eventsrc-example/internal/entity/orders"
)

// GetProjection returns the projection for an order.
// If no events are found, it returns nil, nil.
func (c *Controller) GetProjection(ctx context.Context, orderId string) (*orders.OrderProjection, error) {
	events, err := c.store.ListByAggregateID(ctx, orderId, orders.AggregateTypeOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to list events for order %s: %w", orderId, err)
	}

	if len(events) == 0 {
		return nil, nil
	}

	projEvents := []orders.SerializedEvent{}
	for _, event := range events {
		projEvents = append(projEvents, orders.SerializedEvent{
			EventType: event.EventType,
			EventData: event.Data,
		})
	}

	return orders.ReduceToProjection(projEvents)
}
