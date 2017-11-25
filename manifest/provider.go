package manifest

// Resource provider
type Provider struct {
	Nature string // Resource nature: range, pool ...
	Kind   string // Logical kind
	Config map[string]interface{}
}
