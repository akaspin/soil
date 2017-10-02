package metrics

// BlackHole reporter
type BlackHole struct{}

func (*BlackHole) Count(name string, value int64, tags ...string) {}
