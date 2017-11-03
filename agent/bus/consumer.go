package bus

// Consumer consumes messages
type Consumer interface {
	ConsumeMessage(message Message)
}

type MultiConsumer interface {
	ConsumeMessages(messages ...Message)
}
