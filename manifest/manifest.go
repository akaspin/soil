package manifest

// WithID represents single manifest entity
type WithID interface {
	GetID(parent ...string) string // Get entity ID
}
