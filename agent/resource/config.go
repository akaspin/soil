package resource

// Config represents one resource configuration
type Config struct {
	Type       string
	Name       string
	Properties map[string]interface{}
}
