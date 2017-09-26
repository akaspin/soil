package public

type operator struct {
	backend *Backend
	withTTL bool
	prefix string
}

func NewPermanentOperator(backend *Backend, prefix string) (o *operator) {
	o = &operator{
		backend: backend,
		withTTL: false,
		prefix: prefix,
	}
	return
}

func NewEphemeralOperator(backend *Backend, prefix string) (o *operator) {
	o = &operator{
		backend: backend,
		withTTL: true,
		prefix: prefix,
	}
	return
}

func (o *operator) Set(data map[string]string) {
	o.backend.set(o.prefix, data, o.withTTL)
}

func (o *operator) Delete(key ...string) {
	o.backend.Delete(key...)
}