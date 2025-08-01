package controller

import (
	"context"
	"errors"
	"testing"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	"github.com/cgund98/go-eventsrc-example/internal/entity/orders"
	"github.com/cgund98/go-eventsrc-example/internal/infra/eventsrc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockProducer is a mock implementation of eventsrc.Producer
type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) Send(ctx context.Context, args *eventsrc.SendArgs) error {
	callArgs := m.Called(ctx, args)
	return callArgs.Error(0)
}

func TestController_PlaceOrder(t *testing.T) {
	t.Run("successful order placement", func(t *testing.T) {
		mockProducer := &MockProducer{}
		mockProducer.On("Send", mock.Anything, mock.MatchedBy(func(args *eventsrc.SendArgs) bool {
			return args.AggregateType == orders.AggregateTypeOrder &&
				args.EventType == orders.EventTypeOrderPlaced &&
				args.AggregateID != "" &&
				len(args.Value) > 0
		})).Return(nil)

		controller := &Controller{producer: mockProducer}

		request := &pb.PlaceOrderRequest{
			VendorId:      "vendor-123",
			CustomerId:    "customer-456",
			ProductId:     "product-789",
			Quantity:      2,
			TotalPrice:    99.99,
			PaymentMethod: "credit_card",
		}

		response, err := controller.PlaceOrder(context.Background(), request)

		assert.NoError(t, err)
		require.NotNil(t, response)
		assert.NotEmpty(t, response.OrderId)
		mockProducer.AssertExpectations(t)
	})

	t.Run("producer send error", func(t *testing.T) {
		mockProducer := &MockProducer{}
		mockProducer.On("Send", mock.Anything, mock.Anything).Return(errors.New("database error"))

		controller := &Controller{producer: mockProducer}

		request := &pb.PlaceOrderRequest{
			VendorId:      "vendor-123",
			CustomerId:    "customer-456",
			ProductId:     "product-789",
			Quantity:      1,
			TotalPrice:    50.00,
			PaymentMethod: "debit_card",
		}

		response, err := controller.PlaceOrder(context.Background(), request)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send order placed event")
		assert.Nil(t, response)
		mockProducer.AssertExpectations(t)
	})
}
