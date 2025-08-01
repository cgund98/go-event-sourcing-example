package eventsrc

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockConsumer is a mock implementation of Consumer
type MockConsumer struct {
	mock.Mock
}

func (m *MockConsumer) Name() string {
	return "mock-consumer"
}

func (m *MockConsumer) Consume(ctx context.Context, eventType string, eventData []byte) error {
	callArgs := m.Called(ctx, eventType, eventData)
	return callArgs.Error(0)
}

// MockReader is a mock implementation of Reader
type MockReader struct {
	mock.Mock
}

func (m *MockReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	callArgs := m.Called(ctx)
	return callArgs.Get(0).(kafka.Message), callArgs.Error(1)
}

func (m *MockReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	callArgs := m.Called(ctx, msgs)
	return callArgs.Error(0)
}

func TestRunKafkaConsumerOnce(t *testing.T) {
	t.Run("successful message processing", func(t *testing.T) {
		mockReader := &MockReader{}
		mockConsumer := &MockConsumer{}

		// Create a test message with event type header
		msg := kafka.Message{
			Value: []byte("test event data"),
			Headers: []kafka.Header{
				{Key: KafkaHeaderEventType, Value: []byte("test_event")},
			},
		}

		mockReader.On("FetchMessage", mock.Anything).Return(msg, nil)
		mockConsumer.On("Consume", mock.Anything, "test_event", []byte("test event data")).Return(nil)
		mockReader.On("CommitMessages", mock.Anything, []kafka.Message{msg}).Return(nil)

		err := runKafkaConsumerOnce(context.Background(), mockReader, mockConsumer)

		assert.NoError(t, err)
		mockReader.AssertExpectations(t)
		mockConsumer.AssertExpectations(t)
	})

	t.Run("fetch message error", func(t *testing.T) {
		mockReader := &MockReader{}
		mockConsumer := &MockConsumer{}

		mockReader.On("FetchMessage", mock.Anything).Return(kafka.Message{}, errors.New("kafka error"))

		err := runKafkaConsumerOnce(context.Background(), mockReader, mockConsumer)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "kafka error")
		mockReader.AssertExpectations(t)
		mockConsumer.AssertNotCalled(t, "Consume")
	})

	t.Run("consumer error", func(t *testing.T) {
		mockReader := &MockReader{}
		mockConsumer := &MockConsumer{}

		msg := kafka.Message{
			Value: []byte("test event data"),
			Headers: []kafka.Header{
				{Key: KafkaHeaderEventType, Value: []byte("test_event")},
			},
		}

		mockReader.On("FetchMessage", mock.Anything).Return(msg, nil)
		mockConsumer.On("Consume", mock.Anything, "test_event", []byte("test event data")).Return(errors.New("consumer error"))

		err := runKafkaConsumerOnce(context.Background(), mockReader, mockConsumer)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "consumer error")
		mockReader.AssertExpectations(t)
		mockConsumer.AssertExpectations(t)
		// Should not commit on consumer error
		mockReader.AssertNotCalled(t, "CommitMessages")
	})

	t.Run("commit error", func(t *testing.T) {
		mockReader := &MockReader{}
		mockConsumer := &MockConsumer{}

		msg := kafka.Message{
			Value: []byte("test event data"),
			Headers: []kafka.Header{
				{Key: KafkaHeaderEventType, Value: []byte("test_event")},
			},
		}

		mockReader.On("FetchMessage", mock.Anything).Return(msg, nil)
		mockConsumer.On("Consume", mock.Anything, "test_event", []byte("test event data")).Return(nil)
		mockReader.On("CommitMessages", mock.Anything, []kafka.Message{msg}).Return(errors.New("commit error"))

		err := runKafkaConsumerOnce(context.Background(), mockReader, mockConsumer)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "commit error")
		mockReader.AssertExpectations(t)
		mockConsumer.AssertExpectations(t)
	})

	t.Run("missing event type header", func(t *testing.T) {
		mockReader := &MockReader{}
		mockConsumer := &MockConsumer{}

		msg := kafka.Message{
			Value:   []byte("test event data"),
			Headers: []kafka.Header{}, // No event type header
		}

		mockReader.On("FetchMessage", mock.Anything).Return(msg, nil)
		mockConsumer.On("Consume", mock.Anything, "", []byte("test event data")).Return(nil)
		mockReader.On("CommitMessages", mock.Anything, []kafka.Message{msg}).Return(nil)

		err := runKafkaConsumerOnce(context.Background(), mockReader, mockConsumer)

		assert.NoError(t, err)
		mockReader.AssertExpectations(t)
		mockConsumer.AssertExpectations(t)
	})
}

func TestRunKafkaConsumer(t *testing.T) {
	t.Run("context cancellation", func(t *testing.T) {
		mockReader := &MockReader{}
		mockConsumer := &MockConsumer{}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := RunKafkaConsumer(ctx, mockReader, mockConsumer, RunKafkaConsumerOptions{})

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		mockReader.AssertNotCalled(t, "FetchMessage")
	})
}
