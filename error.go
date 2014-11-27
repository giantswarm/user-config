package userconfig

import (
	"github.com/juju/errgo"
)

var (
	ErrUnknownJSONField = errgo.New("Unknown JSON field.")

	Mask = errgo.MaskFunc(IsErrUnknownJsonField)
)

func IsErrUnknownJsonField(err error) bool {
	return errgo.Cause(err) == ErrUnknownJSONField
}
