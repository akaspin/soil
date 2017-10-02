package metrics

type Reporter interface {
	Count(name string, value int64, tags ...string)
}
