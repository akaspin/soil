package backend

type operator struct {
	backend *LibKVBackend
	withTTL bool
	prefix  string
}

func NewPermanentOperator(backend *LibKVBackend, prefix string) (o *operator) {
	o = &operator{
		backend: backend,
		withTTL: false,
		prefix:  prefix,
	}
	return
}

func NewEphemeralOperator(backend *LibKVBackend, prefix string) (o *operator) {
	o = &operator{
		backend: backend,
		withTTL: true,
		prefix:  prefix,
	}
	return
}

func (o *operator) Set(data map[string]string) {
	o.backend.set(o.prefix, data, o.withTTL)
}

func (o *operator) Delete(key ...string) {
	o.backend.deleteWithPrefix(o.prefix, key...)
}
