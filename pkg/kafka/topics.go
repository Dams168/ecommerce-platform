package kafka

// Semua topik di kafka import konstanta ini, tidak boleh hardcode string
// typo topic = pesan tidak sampai = bug susah dicari
const (
    TopicOrderCreated = "order.created"
    TopicPaymentDone = "payment.done"
    TopicPaymentFailed = "payment.failed"
)

// Consumer group - satu group = satu tim worker
// kafka membagi partisi secara merata ke semua member group
const (
    GroupPaymentworker = "payment-workers"
    GroupNotificationSvc = "notification-service"
)