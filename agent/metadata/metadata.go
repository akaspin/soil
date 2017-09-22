package metadata

type Message struct {

	// Producer clean status
	Clean bool

	// Producer prefix
	Prefix string

	// Message payload
	Data map[string]string
}

// Producer syncs changed data to consumers
type Producer interface {

	// Prefix returns source prefix
	Prefix() string

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