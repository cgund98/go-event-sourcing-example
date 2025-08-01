package eventsrc

import (
	"context"
	"time"

	"github.com/cgund98/go-eventsrc-example/internal/infra/logging"

	"github.com/segmentio/kafka-go"
)

const ConsumerRetryDelay = 5 * time.Second

// Consumer is an interface for consuming events from the event message bus.
type Consumer interface {
	Name() string
	Consume(ctx context.Context, eventType string, eventData []byte) error
}

// Reader is an interface for reading messages from Kafka.
type Reader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
}

func parseEventType(msg kafka.Message) string {
	for _, header := range msg.Headers {
		if header.Key == KafkaHeaderEventType {
			return string(header.Value)
		}
	}
	return ""
}

// runKafkaConsumerOnce reads a single event from the reader and passes it to the consumer.
func runKafkaConsumerOnce(ctx context.Context, reader Reader, consumer Consumer) error {
	event, err := reader.FetchMessage(ctx)
	if err != nil {
		logging.Logger.Error("error reading event", "error", err)
		return err
	}

	logging.Logger.Debug("Received event", "eventType", parseEventType(event), "consumer", consumer.Name())

	eventType := parseEventType(event)

	err = consumer.Consume(ctx, eventType, event.Value)
	if err != nil {
		logging.Logger.Error("error consuming event", "error", err)
		return err
	}

	err = reader.CommitMessages(ctx, event)
	if err != nil {
		logging.Logger.Error("error committing event", "error", err)
		return err
	}

	return nil
}

type RunKafkaConsumerOptions struct {
	RetryDelay *time.Duration
}

// RunConsumer runs a kafka consumer in a loop.
// It will read events from the reader and pass them to the consumer.
func RunKafkaConsumer(ctx context.Context, reader Reader, consumer Consumer, opts RunKafkaConsumerOptions) error {

	// Parse options
	retryDelay := ConsumerRetryDelay
	if opts.RetryDelay != nil {
		retryDelay = *opts.RetryDelay
	}

	logging.Logger.Info("Starting kafka consumer", "consumer", consumer.Name())

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := runKafkaConsumerOnce(ctx, reader, consumer)
			if err != nil {
				logging.Logger.Error("error running kafka consumer", "error", err)
				time.Sleep(retryDelay)
			}
		}
	}
}
