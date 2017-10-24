package bus

// BindableProducer after its Open
type BindableProducer interface {

	// RegisterConsumer registers consumer on BindableProducer
	RegisterConsumer(name string, consumer Consumer)
}

// Consumer consumes messages
type Consumer interface {
	ConsumeMessage(message Message)
}
