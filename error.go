package userconfig

import (
	"encoding/json"
	"github.com/juju/errgo"
)

var (
	UnknownJSONFieldError        = errgo.New("Unknown JSON field.")
	InvalidSizeError             = errgo.New("Invalid size.")
	DuplicateVolumePathError     = errgo.New("Duplicate volume path.")
	InvalidEnvListFormatError    = errgo.Newf("Unable to parse 'env'. Objects or Array of strings expected.")
	CrossServicePodError         = errgo.New("Pod is used in different services.")
	PodUsedOnlyOnceError         = errgo.New("Pod is used in only 1 component.")
	InvalidVolumeConfigError     = errgo.New("Invalid volume configuration.")
	InvalidDependencyConfigError = errgo.New("Invalid dependency configuration.")
	InvalidScalingConfigError    = errgo.New("Invalid scaling configuration.")

	Mask = errgo.MaskFunc(IsInvalidEnvListFormat,
		IsUnknownJsonField,
		IsInvalidSize,
		IsDuplicateVolumePath,
		IsCrossServicePod,
		IsPodUsedOnlyOnce,
		IsInvalidVolumeConfig,
		IsInvalidDependencyConfig,
		IsInvalidScalingConfig,
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

func IsCrossServicePod(err error) bool {
	return errgo.Cause(err) == CrossServicePodError
}

func IsPodUsedOnlyOnce(err error) bool {
	return errgo.Cause(err) == PodUsedOnlyOnceError
}

func IsInvalidVolumeConfig(err error) bool {
	return errgo.Cause(err) == InvalidVolumeConfigError
}

func IsInvalidDependencyConfig(err error) bool {
	return errgo.Cause(err) == InvalidDependencyConfigError
}

func IsInvalidScalingConfig(err error) bool {
	return errgo.Cause(err) == InvalidScalingConfigError
}

// IsSyntax returns true if the cause of the given error in a json.SyntaxError
func IsSyntax(err error) bool {
	_, ok := errgo.Cause(err).(*json.SyntaxError)
	return ok
}
