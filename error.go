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
	InvalidPortConfigError       = errgo.New("Invalid port configuration.")
	InvalidDomainDefintionError  = errgo.New("invalid domain definition")
	InvalidLinkDefinitionError   = errgo.New("invalid link definition")
	InvalidAppDefinitionError    = errgo.New("invalid app definition")
	InvalidNodeDefinitionError   = errgo.New("invalid node definition")
	InvalidImageDefinitionError  = errgo.New("invalid image definition")
	InvalidNodeNameError         = errgo.New("invalid node name")
	NodeNotFoundError            = errgo.New("node not found")

	Mask = errgo.MaskFunc(IsInvalidEnvListFormat,
		IsUnknownJsonField,
		IsInvalidSize,
		IsDuplicateVolumePath,
		IsCrossServicePod,
		IsPodUsedOnlyOnce,
		IsInvalidVolumeConfig,
		IsInvalidDependencyConfig,
		IsInvalidScalingConfig,
		IsInvalidPortConfig,
		IsInvalidDomainDefinition,
		IsInvalidLinkDefinition,
		IsInvalidAppDefinition,
		IsInvalidNodeDefinition,
		IsInvalidImageDefinition,
		IsInvalidNodeName,
		IsNodeNotFound,
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

func IsInvalidPortConfig(err error) bool {
	return errgo.Cause(err) == InvalidPortConfigError
}

func IsInvalidDomainDefinition(err error) bool {
	return errgo.Cause(err) == InvalidDomainDefintionError
}

func IsInvalidLinkDefinition(err error) bool {
	return errgo.Cause(err) == InvalidLinkDefinitionError
}

func IsInvalidAppDefinition(err error) bool {
	return errgo.Cause(err) == InvalidAppDefinitionError
}

func IsInvalidNodeDefinition(err error) bool {
	return errgo.Cause(err) == InvalidNodeDefinitionError
}

func IsInvalidImageDefinition(err error) bool {
	return errgo.Cause(err) == InvalidImageDefinitionError
}

func IsInvalidNodeName(err error) bool {
	return errgo.Cause(err) == InvalidNodeNameError
}

func IsNodeNotFound(err error) bool {
	return errgo.Cause(err) == NodeNotFoundError
}

// IsSyntax returns true if the cause of the given error in a json.SyntaxError
func IsSyntax(err error) bool {
	_, ok := errgo.Cause(err).(*json.SyntaxError)
	return ok
}
