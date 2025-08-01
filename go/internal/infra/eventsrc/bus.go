package eventsrc

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

const KafkaWriteTimeout = 10 * time.Second
const KafkaHeaderAggregateID = "aggregate-id"
const KafkaHeaderAggregateType = "aggregate-type"
const KafkaHeaderEventType = "event-type"

type Bus interface {
	Publish(ctx context.Context, args *PublishArgs) error
}

type PublishArgs struct {
	AggregateID   string
	AggregateType string
	EventType     string
	Value         []byte
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
			{
				Key:   KafkaHeaderAggregateID,
				Value: []byte(args.AggregateID),
			},
			{
				Key:   KafkaHeaderAggregateType,
				Value: []byte(args.AggregateType),
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

func GetAggregateIDFromMessage(msg *kafka.Message) (string, error) {
	for _, header := range msg.Headers {
		if header.Key == KafkaHeaderAggregateID {
			return string(header.Value), nil
		}
	}
	return "", fmt.Errorf("aggregate id not found in message")
}

func GetAggregateTypeFromMessage(msg *kafka.Message) (string, error) {
	for _, header := range msg.Headers {
		if header.Key == KafkaHeaderAggregateType {
			return string(header.Value), nil
		}
	}
	return "", fmt.Errorf("aggregate type not found in message")
}

func GetEventTypeFromMessage(msg *kafka.Message) (string, error) {
	for _, header := range msg.Headers {
		if header.Key == KafkaHeaderEventType {
			return string(header.Value), nil
		}
	}
	return "", fmt.Errorf("event type not found in message")
}

/** In Memory Bus */
type BusEvent struct {
	EventType string
	Data      []byte
}

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
