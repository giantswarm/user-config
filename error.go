package userconfig

import (
	"github.com/juju/errgo"
)

var (
	ErrUnknownJSONField = errgo.New("Unknown JSON field.")
	ErrInvalidSize      = errgo.New("Invalid size.")

	Mask = errgo.MaskFunc(IsErrUnknownJsonField, IsErrInvalidSize)
)

func IsErrUnknownJsonField(err error) bool {
	return errgo.Cause(err) == ErrUnknownJSONField
}

func IsErrInvalidSize(err error) bool {
	return errgo.Cause(err) == ErrInvalidSize
}
