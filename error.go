package userconfig

import (
	"encoding/json"
	"github.com/juju/errgo"
)

var (
	UnknownJSONFieldError         = errgo.New("Unknown JSON field.")
	MissingJSONFieldError         = errgo.New("missing JSON field")
	InvalidSizeError              = errgo.New("Invalid size.")
	DuplicateVolumePathError      = errgo.New("Duplicate volume path.")
	InvalidEnvListFormatError     = errgo.Newf("Unable to parse 'env'. Objects or Array of strings expected.")
	CrossServicePodError          = errgo.New("Pod is used in different services.")
	PodUsedOnlyOnceError          = errgo.New("Pod is used in only 1 component.")
	InvalidVolumeConfigError      = errgo.New("Invalid volume configuration.")
	InvalidDependencyConfigError  = errgo.New("Invalid dependency configuration.")
	InvalidScalingConfigError     = errgo.New("Invalid scaling configuration.")
	InvalidPortConfigError        = errgo.New("Invalid port configuration.")
	InvalidDomainDefinitionError  = errgo.New("invalid domain definition")
	InvalidLinkDefinitionError    = errgo.New("invalid link definition")
	InvalidAppDefinitionError     = errgo.New("invalid app definition")
	InvalidNodeDefinitionError    = errgo.New("invalid node definition")
	InvalidImageDefinitionError   = errgo.New("invalid image definition")
	InvalidNodeNameError          = errgo.New("invalid node name")
	NodeNotFoundError             = errgo.New("node not found")
	InternalError                 = errgo.New("internal error")
	MissingValidationContextError = errgo.New("missing validation context")

	mask = errgo.MaskFunc(IsInvalidEnvListFormat,
		IsUnknownJsonField,
		IsMissingJsonField,
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
		IsInternal,
		IsMissingValidationContext,
	)
)

// maskf is short for mask(errgo.WithCausef(nil, cause, f, a...))
func maskf(cause error, f string, a ...interface{}) error {
	err := mask(errgo.WithCausef(nil, cause, f, a...))
	// the above call with set this location instead of that of our caller.
	// that is fixed below.
	if e, _ := err.(*errgo.Err); e != nil {
		e.SetLocation(1)
	}
	return err
}

func IsUnknownJsonField(err error) bool {
	return errgo.Cause(err) == UnknownJSONFieldError
}

func IsMissingJsonField(err error) bool {
	return errgo.Cause(err) == MissingJSONFieldError
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
	return errgo.Cause(err) == InvalidDomainDefinitionError
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

func IsInternal(err error) bool {
	return errgo.Cause(err) == InternalError
}

func IsMissingValidationContext(err error) bool {
	return errgo.Cause(err) == MissingValidationContextError
}

// IsSyntax returns true if the cause of the given error in a json.SyntaxError
func IsSyntax(err error) bool {
	_, ok := errgo.Cause(err).(*json.SyntaxError)
	return ok
}
