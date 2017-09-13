package agent

// SourceProducer syncs changed data to consumers
type SourceProducer interface {

	// Prefix returns source prefix
	Prefix() string

	// Register consumer in source
	RegisterConsumer(name string, consumer SourceConsumer)

	// Notify registered consumers
	Notify()
}

type SourceConsumer interface {

	// Sync called by Source producer on data change
	Sync(producer string, active bool, data map[string]string)
}
