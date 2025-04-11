package object

import "errors"

var (
	ErrTypeAssertion = errors.New("type assertion failed")
	ErrBase64Decode2 = errors.New("unrecognized level")
)
