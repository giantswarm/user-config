package userconfig

import (
	"encoding/json"
	"github.com/juju/errgo"
)

var (
	UnknownJSONFieldError           = errgo.New("unknown JSON field")
	MissingJSONFieldError           = errgo.New("missing JSON field")
	InvalidSizeError                = errgo.New("invalid size")
	DuplicateVolumePathError        = errgo.New("duplicate volume path")
	InvalidEnvListFormatError       = errgo.Newf("unable to parse 'env', objects or Array of strings expected")
	CrossServicePodError            = errgo.New("pod is used in different services")
	PodUsedOnlyOnceError            = errgo.New("pod is used in only 1 component")
	InvalidVolumeConfigError        = errgo.New("invalid volume configuration")
	InvalidDependencyConfigError    = errgo.New("invalid dependency configuration")
	InvalidScalingConfigError       = errgo.New("invalid scaling configuration")
	InvalidPortConfigError          = errgo.New("Invalid port configuration")
	InvalidDomainDefinitionError    = errgo.New("invalid domain definition")
	InvalidLinkDefinitionError      = errgo.New("invalid link definition")
	InvalidAppDefinitionError       = errgo.New("invalid service definition")
	InvalidComponentDefinitionError = errgo.New("invalid component definition")
	InvalidImageDefinitionError     = errgo.New("invalid image definition")
	InvalidAppNameError             = errgo.New("invalid service name")
	InvalidComponentNameError       = errgo.New("invalid component name")
	InvalidPodConfigError           = errgo.New("invalid pod configuration")
	PortNotFoundError               = errgo.New("port not found")
	ComponentNotFoundError          = errgo.New("component not found")
	InternalError                   = errgo.New("internal error")
	MissingValidationContextError   = errgo.New("missing validation context")
	InvalidArgumentError            = errgo.New("invalid argument")
	VolumeCycleError                = errgo.New("cycle detected in volume configuration")
	WrongDiffOrderError             = errgo.New("wrong diff order")
	LinkCycleError                  = errgo.New("cycle detected in link definition")

	mask = errgo.MaskFunc(IsInvalidEnvListFormat,
		IsUnknownJsonField,
		IsMissingJsonField,
		IsInvalidSize,
		IsDuplicateVolumePath,
		IsCrossServicePod,
		IsPodUsedOnlyOnce,
		IsInvalidVolumeConfig,
		IsVolumeCycle,
		IsInvalidDependencyConfig,
		IsInvalidScalingConfig,
		IsInvalidPortConfig,
		IsInvalidPodConfig,
		IsInvalidDomainDefinition,
		IsInvalidLinkDefinition,
		IsInvalidAppDefinition,
		IsInvalidComponentDefinition,
		IsInvalidImageDefinition,
		IsInvalidAppName,
		IsInvalidComponentName,
		IsComponentNotFound,
		IsPortNotFound,
		IsInternal,
		IsMissingValidationContext,
		IsInvalidArgument,
		IsSyntax,
		IsLinkCycle,
	)

	maskAny = errgo.MaskFunc(errgo.Any)
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

// IsInvalidVolumeConfig returns true if the given error is of type
// InvalidVolumeConfigError, VolumeCycleError or DuplicateVolumePathError. False otherwise.
func IsInvalidVolumeConfig(err error) bool {
	cause := errgo.Cause(err)
	return cause == InvalidVolumeConfigError || cause == DuplicateVolumePathError || cause == VolumeCycleError
}

func IsVolumeCycle(err error) bool {
	return errgo.Cause(err) == VolumeCycleError
}

func IsInvalidDependencyConfig(err error) bool {
	return errgo.Cause(err) == InvalidDependencyConfigError
}

func IsInvalidScalingConfig(err error) bool {
	return errgo.Cause(err) == InvalidScalingConfigError
}

func IsInvalidPodConfig(err error) bool {
	return errgo.Cause(err) == InvalidPodConfigError
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

func IsInvalidComponentDefinition(err error) bool {
	return errgo.Cause(err) == InvalidComponentDefinitionError
}

func IsInvalidImageDefinition(err error) bool {
	return errgo.Cause(err) == InvalidImageDefinitionError
}

func IsInvalidAppName(err error) bool {
	return errgo.Cause(err) == InvalidAppNameError
}

func IsInvalidComponentName(err error) bool {
	return errgo.Cause(err) == InvalidComponentNameError
}

func IsComponentNotFound(err error) bool {
	return errgo.Cause(err) == ComponentNotFoundError
}

func IsPortNotFound(err error) bool {
	return errgo.Cause(err) == PortNotFoundError
}

func IsInternal(err error) bool {
	return errgo.Cause(err) == InternalError
}

func IsMissingValidationContext(err error) bool {
	return errgo.Cause(err) == MissingValidationContextError
}

func IsInvalidArgument(err error) bool {
	return errgo.Cause(err) == InvalidArgumentError
}

func IsLinkCycle(err error) bool {
	return errgo.Cause(err) == LinkCycleError
}

// IsSyntax returns true if the cause of the given error in a json.SyntaxError
func IsSyntax(err error) bool {
	_, ok := errgo.Cause(err).(*json.SyntaxError)
	return ok
}
