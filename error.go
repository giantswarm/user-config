package userconfig

import (
	"github.com/juju/errgo"
)

var (
	ErrUnknownJSONField     = errgo.New("Unknown JSON field.")
	ErrInvalidSize          = errgo.New("Invalid size.")
	ErrDuplicateVolumePath  = errgo.New("Duplicate volume path.")
	ErrInvalidEnvListFormat = errgo.Newf("Unable to parse 'env'. Objects or Array of strings expected.")

	Mask = errgo.MaskFunc(IsErrInvalidEnvListFormat, IsErrUnknownJsonField, IsErrInvalidSize, IsErrDuplicateVolumePath)
)

func IsErrUnknownJsonField(err error) bool {
	return errgo.Cause(err) == ErrUnknownJSONField
}

func IsErrInvalidSize(err error) bool {
	return errgo.Cause(err) == ErrInvalidSize
}

func IsErrDuplicateVolumePath(err error) bool {
	return errgo.Cause(err) == ErrDuplicateVolumePath
}

func IsErrInvalidEnvListFormat(err error) bool {
	return errgo.Cause(err) == ErrInvalidEnvListFormat
}
