package bus

// Pipe consumes messages and pipes them to downstream
type Pipe interface {
	GetConsumer() (c Consumer)
}

// Replicated pipes messages to multiple pipes
type ReplicatedPipe interface {
	GetConsumers() (c []Consumer)
}
