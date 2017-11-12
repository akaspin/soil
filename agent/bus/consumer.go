package bus

// Consumer consumes messages
type Consumer interface {
	ConsumeMessage(message Message)
}

type NamedConsumer interface {
	ConsumerName() string
}
