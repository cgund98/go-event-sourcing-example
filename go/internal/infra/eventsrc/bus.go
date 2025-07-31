package eventsrc

import (
	"context"

	"github.com/segmentio/kafka-go"
)

const KafkaHeaderEventType = "event-type"

type Bus interface {
	Publish(ctx context.Context, args *PublishArgs) error
}

type PublishArgs struct {
	EventType string
	Value     []byte
}

type BusEvent struct {
	EventType string
	Data      []byte
}

/** Kafka Bus */

type KafkaBus struct {
	conn *kafka.Conn
}

func NewKafkaBus(conn *kafka.Conn) *KafkaBus {
	return &KafkaBus{conn: conn}
}

func (b *KafkaBus) Publish(ctx context.Context, args *PublishArgs) error {
	msg := kafka.Message{
		Headers: []kafka.Header{
			{
				Key:   KafkaHeaderEventType,
				Value: []byte(args.EventType),
			},
		},
		Value: args.Value,
	}

	_, err := b.conn.WriteMessages(msg)
	if err != nil {
		return err
	}

	return err
}

/** In Memory Bus */

type InMemoryBus struct {
	Events []BusEvent
}

func NewInMemoryBus() *InMemoryBus {
	return &InMemoryBus{Events: []BusEvent{}}
}

func (p *InMemoryBus) Publish(ctx context.Context, args *PublishArgs) error {
	p.Events = append(p.Events, BusEvent{
		EventType: args.EventType,
		Data:      args.Value,
	})
	return nil
}
