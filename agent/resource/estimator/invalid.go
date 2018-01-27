package estimator

// Invalid estimator accepts resource requests but always responds that
// all of them are failed
type Invalid struct {
	*base
}

func NewInvalid(globalConfig GlobalConfig, config Config) (i *Invalid) {
	i = &Invalid{}
	i.base = newBase(globalConfig, config, i)
	return i
}

func (i *Invalid) createFn(id string, config map[string]interface{}, values map[string]string) (res interface{}, err error) {
	err = ErrInvalidProvider
	i.send(id, err, nil)
	return nil, err
}

func (i *Invalid) updateFn(id string, config map[string]interface{}) (res interface{}, err error) {
	err = ErrInvalidProvider
	i.send(id, err, nil)
	return nil, err
}

func (i *Invalid) destroyFn(id string) (err error) {
	i.send(id, nil, nil)
	return nil
}

func (i *Invalid) shutdownFn() (err error) {
	return nil
}
