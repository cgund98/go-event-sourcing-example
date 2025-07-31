package eventsrc

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

const KafkaWriteTimeout = 10 * time.Second
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
	writer *kafka.Writer
}

func NewKafkaBus(writer *kafka.Writer) *KafkaBus {
	return &KafkaBus{writer: writer}
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

	wCtx, cancel := context.WithTimeout(ctx, KafkaWriteTimeout)
	defer cancel()

	err := b.writer.WriteMessages(wCtx, msg)
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
