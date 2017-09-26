package bus

type Setter interface {
	Set(data map[string]string)
}

type Deleter interface {
	Delete(key ...string)
}