package estimator

const (
	BlackholeEstimator = "blackhole" // blackhole estimator name
)

// Blackhole estimator accepts all requests but evaluates only destroys
type Blackhole struct {
	*base
}

func NewBlackhole(globalConfig GlobalConfig, config Config) (b *Blackhole) {
	b = &Blackhole{}
	b.base = newBase(globalConfig, config, b)
	return
}

func (b *Blackhole) createFn(id string, config map[string]interface{}, values map[string]string) (res interface{}, err error) {
	return
}

func (b *Blackhole) updateFn(id string, config map[string]interface{}) (res interface{}, err error) {
	return
}

func (b *Blackhole) destroyFn(id string) (err error) {
	b.send(id, nil, nil)
	return
}

func (b *Blackhole) shutdownFn() (err error) {

	return
}
