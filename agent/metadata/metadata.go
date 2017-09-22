package metadata

type Message struct {

	// DynamicProducer clean status
	Clean bool

	// DynamicProducer prefix
	Prefix string

	// Message payload
	Data map[string]string
}

// DynamicProducer permits to add consumers after initialisation
type DynamicProducer interface {

	// Register consumer in source
	RegisterConsumer(name string, fn func(message Message))
}

type Consumer interface {

	// Sync called by Source producer on data change
	Sync(message Message)
}

type Upstream interface {
	Replace(data map[string]string)
	Set(data map[string]string)
	Delete(keys ...string)
}
