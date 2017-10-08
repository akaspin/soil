package bus

type Setter interface {
	Set(data map[string]string)
}

type Deleter interface {
	Delete(key ...string)
}

type Upstream interface {
	Setter
	Deleter
}
