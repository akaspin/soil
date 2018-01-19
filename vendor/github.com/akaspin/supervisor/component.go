package supervisor

// Component is basic building block to build supervisor trees
type Component interface {

	// Open runs Component initialisation and should blocks until
	// Component is initialised.
	Open() (err error)

	// Close initialises Component shutdown.
	Close() (err error)

	// Wait should blocks until Component shutdown.
	Wait() (err error)
}
