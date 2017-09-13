package agent


type Source interface {
	SourceProducer

	// pod namespaces managed by source
	Namespaces() []string

	// Is data used only in constraint or
	// available for interpolation
	Mark() bool
}

// SourceProducer syncs changed data to consumers
type SourceProducer interface {

	// Prefix returns source prefix
	Prefix() string

	// Register consumer in source
	RegisterConsumer(name string, consumer SourceConsumer)
}

type SourceConsumer interface {

	// Sync called by Source producer on data change
	Sync(producer string, active bool, data map[string]string)
}
