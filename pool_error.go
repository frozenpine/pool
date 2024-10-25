package pool

import "errors"

var (
	ErrInvalidType = errors.New("generic type must be a struct")
)
