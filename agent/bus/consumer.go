package bus

// Consumer consumes messages
type Consumer interface {
	ConsumeMessage(message Message) (err error)
}
