package estimator

import "errors"

var (
	ErrNotAvailable    = errors.New("not-available")
	ErrInvalidProvider = errors.New("invalid-provider-kind")
)
