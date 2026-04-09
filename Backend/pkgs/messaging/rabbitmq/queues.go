package rmq_config

const (
	ExchangeBooking = "cloud.booking.topic"
	ExchangeDLX     = "cloud.booking.dlx"

	RoutingRefundRequest = "booking.refund.requested"
	RoutingRefundApprove = "booking.refund.approved"
	RoutingRefundReject  = "booking.refund.rejected"
	RoutingDLXDead       = "booking.events.dead"

	QueueAdminRefund    = "cloud.admin.refund.queue"
	QueueAdminRefundDLQ = "cloud.admin.refund.dlq"
)
