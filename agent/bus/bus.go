package bus

import "github.com/akaspin/soil/manifest"

// BindableProducer after its Open
type BindableProducer interface {

	// RegisterConsumer registers consumer on BindableProducer
	RegisterConsumer(name string, consumer MessageConsumer)
}

// MessageConsumer consumes messages
type MessageConsumer interface {

	// ConsumeMessage called by producer on data change
	ConsumeMessage(message Message)
}

type RegistryConsumer interface {
	ConsumeRegistry(namespace string, payload manifest.Registry)
}
