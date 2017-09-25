package metadata

// DynamicProducer permits to add consumers after initialisation
type DynamicProducer interface {

	// Register consumer in source
	RegisterConsumer(name string, consumer Consumer)
}

type Consumer interface {

	// ConsumeMessage called by Source producer on data change
	ConsumeMessage(message Message)
}

type Upstream interface {
	Set(map[string]string)
	Replace(map[string]string)
}
