package orders

const (
	EventTypeOrderPlaced                = "order_placed"
	EventTypeOrderPaid                  = "order_paid"
	EventTypeOrderPaymentInitiated      = "order_payment_initiated"
	EventTypeOrderPaymentFailed         = "order_payment_failed"
	EventTypeOrderCancelled             = "order_cancelled"
	EventTypeOrderShippingStatusUpdated = "order_shipping_status_updated"

	AggregateTypeOrder = "order"
)
