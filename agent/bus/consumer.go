package bus

type Producer interface {

	// Register consumer on Producer
	RegisterConsumer(name string, consumer Consumer)

	// Unregister consumer from Producer
	UnregisterConsumer(name string)
}

// Consumer consumes messages
type Consumer interface {
	ConsumeMessage(message Message)
}
