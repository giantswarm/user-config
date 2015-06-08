package userconfig

import (
	"encoding/json"
	"github.com/juju/errgo"
)

var (
	UnknownJSONFieldError      = errgo.New("Unknown JSON field.")
	InvalidSizeError           = errgo.New("Invalid size.")
	DuplicateVolumePathError   = errgo.New("Duplicate volume path.")
	InvalidEnvListFormatError  = errgo.Newf("Unable to parse 'env'. Objects or Array of strings expected.")
	CrossServiceNamespaceError = errgo.New("Namespace is used in different services.")
	NamespaceUsedOnlyOnceError = errgo.New("Namespace is used in only 1 component.")
	InvalidVolumeConfigError   = errgo.New("Invalid volume configuration.")

	Mask = errgo.MaskFunc(IsInvalidEnvListFormat,
		IsUnknownJsonField,
		IsInvalidSize,
		IsDuplicateVolumePath,
		IsCrossServiceNamespace,
		IsNamespaceUsedOnlyOnce,
		IsInvalidVolumeConfig,
	)
)

func IsUnknownJsonField(err error) bool {
	return errgo.Cause(err) == UnknownJSONFieldError
}

func IsInvalidSize(err error) bool {
	return errgo.Cause(err) == InvalidSizeError
}

func IsDuplicateVolumePath(err error) bool {
	return errgo.Cause(err) == DuplicateVolumePathError
}

func IsInvalidEnvListFormat(err error) bool {
	return errgo.Cause(err) == InvalidEnvListFormatError
}

func IsCrossServiceNamespace(err error) bool {
	return errgo.Cause(err) == CrossServiceNamespaceError
}

func IsNamespaceUsedOnlyOnce(err error) bool {
	return errgo.Cause(err) == NamespaceUsedOnlyOnceError
}

func IsInvalidVolumeConfig(err error) bool {
	return errgo.Cause(err) == InvalidVolumeConfigError
}

// IsSyntax returns true if the cause of the given error in a json.SyntaxError
func IsSyntax(err error) bool {
	_, ok := errgo.Cause(err).(*json.SyntaxError)
	return ok
}
