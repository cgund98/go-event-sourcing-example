package controller

import (
	"context"
	"fmt"

	"github.com/cgund98/go-eventsrc-example/internal/entity/orders"
)

// GetProjection returns the projection for an order.
// If no events are found, it returns nil, nil.
func (c *Controller) GetProjection(ctx context.Context, orderId string) (*orders.OrderProjection, int, error) {
	events, err := c.store.ListByAggregateID(ctx, orderId, orders.AggregateTypeOrder)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list events for order %s: %w", orderId, err)
	}

	if len(events) == 0 {
		return nil, 0, nil
	}

	projEvents := []orders.SerializedEvent{}
	currentSequenceNumber := 0
	for _, event := range events {
		projEvents = append(projEvents, orders.SerializedEvent{
			EventType: event.EventType,
			EventData: event.Data,
		})
		currentSequenceNumber = max(currentSequenceNumber, event.SequenceNumber)
	}

	projection, err := orders.ReduceToProjection(projEvents)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to reduce to projection: %w", err)
	}

	return projection, currentSequenceNumber, nil
}
