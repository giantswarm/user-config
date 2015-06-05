package userconfig

import (
	"encoding/json"
	"github.com/juju/errgo"
)

var (
	ErrUnknownJSONField      = errgo.New("Unknown JSON field.")
	ErrInvalidSize           = errgo.New("Invalid size.")
	ErrDuplicateVolumePath   = errgo.New("Duplicate volume path.")
	ErrInvalidEnvListFormat  = errgo.Newf("Unable to parse 'env'. Objects or Array of strings expected.")
	ErrCrossServiceNamespace = errgo.New("Namespace is used in different services.")
	ErrNamespaceUsedOnlyOnce = errgo.New("Namespace is used in only 1 component.")
	ErrInvalidVolumeConfig   = errgo.New("Invalid volume configuration.")

	Mask = errgo.MaskFunc(IsErrInvalidEnvListFormat,
		IsErrUnknownJsonField,
		IsErrInvalidSize,
		IsErrDuplicateVolumePath,
		IsErrCrossServiceNamespace,
		IsErrNamespaceUsedOnlyOnce,
		IsErrInvalidVolumeConfig,
	)
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

func IsErrCrossServiceNamespace(err error) bool {
	return errgo.Cause(err) == ErrCrossServiceNamespace
}

func IsErrNamespaceUsedOnlyOnce(err error) bool {
	return errgo.Cause(err) == ErrNamespaceUsedOnlyOnce
}

func IsErrInvalidVolumeConfig(err error) bool {
	return errgo.Cause(err) == ErrInvalidVolumeConfig
}

// IsSyntaxError returns true if the cause of the given error in a json.SyntaxError
func IsSyntaxError(err error) bool {
	_, ok := errgo.Cause(err).(*json.SyntaxError)
	return ok
}
