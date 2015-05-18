package userconfig

import (
	"encoding/json"
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

// IsSyntaxError returns true if the cause of the given error in a json.SyntaxError
func IsSyntaxError(err error) bool {
	_, ok := errgo.Cause(err).(*json.SyntaxError)
	return ok
}
