package userconfig

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"
)

type ServiceDefinition struct {
	// Optional service name
	ServiceName ServiceName `json:"name,omitempty"`

	// Components
	Components ComponentDefinitions `json:"components"`
}

// ParseServiceDefinition tries to parse the v2 service definition.
func ParseServiceDefinition(b []byte) (ServiceDefinition, error) {
	var serviceDef ServiceDefinition
	if err := json.Unmarshal(b, &serviceDef); err != nil {
		if IsSyntax(err) {
			if strings.Contains(err.Error(), "$") {
				return ServiceDefinition{}, maskf(err, "Cannot parse swarm.json. Maybe not all variables replaced properly.")
			}
		}

		return ServiceDefinition{}, mask(err)
	}

	return serviceDef, nil
}

func (sd *ServiceDefinition) UnmarshalJSON(data []byte) error {
	// We fix the json buffer so CheckForUnknownFields doesn't complain about
	// `Components` (with uper N).
	data, err := FixJSONFieldNames(data)
	if err != nil {
		return err
	}

	if err := CheckForUnknownFields(data, sd); err != nil {
		return err
	}

	// Just unmarshal the given bytes into the service def struct, since there
	// were no errors.
	var sdc serviceDefCopy
	if err := json.Unmarshal(data, &sdc); err != nil {
		return mask(err)
	}

	result := ServiceDefinition(sdc)

	*sd = result

	return nil
}

type ValidationContext struct {
	Org       string
	Protocols []string

	MinScaleSize int
	MaxScaleSize int
	Placement    Placement

	MinVolumeSize VolumeSize
	MaxVolumeSize VolumeSize

	EnableUserMemoryLimit bool // If false, the component definition MUST NOT have a memory-limit configured
	MinMemoryLimit        ByteSize
	MaxMemoryLimit        ByteSize

	PublicDockerRegistry  string
	PrivateDockerRegistry string
}

// validate performs semantic validations of this ServiceDefinition.
// Return the first possible error.
func (sd *ServiceDefinition) Validate(valCtx *ValidationContext) error {
	if len(sd.Components) == 0 {
		return maskf(InvalidAppDefinitionError, "components must not be empty")
	}

	if !sd.ServiceName.Empty() {
		if err := sd.ServiceName.Validate(); err != nil {
			return mask(err)
		}
	}

	if err := sd.Components.validate(valCtx); err != nil {
		return mask(err)
	}

	return nil
}

// HideDefaults uses the given validation context to determine what definition
// details should be hidden. The caller can clean the definition that way to
// not confuse the user with information he has not set by himself.
func (sd *ServiceDefinition) HideDefaults(valCtx *ValidationContext) (*ServiceDefinition, error) {
	if valCtx == nil {
		return &ServiceDefinition{}, maskf(MissingValidationContextError, "cannot hide defaults")
	}

	sd.Components = sd.Components.hideDefaults(valCtx)
	return sd, nil
}

// SetDefaults sets necessary default values if not given by the user.
func (sd *ServiceDefinition) SetDefaults(valCtx *ValidationContext) error {
	if valCtx == nil {
		return maskf(MissingValidationContextError, "cannot set defaults")
	}

	sd.Components.setDefaults(valCtx)
	return nil
}

// Name returns the name of the given definition if it exists.
// It is does not exist, it generates an service name.
func (sd *ServiceDefinition) Name() (string, error) {
	// Is a name specified?
	if !sd.ServiceName.Empty() {
		return sd.ServiceName.String(), nil
	}

	// No name is specified, generate one
	if name, err := sd.generateServiceName(); err != nil {
		return "", mask(err)
	} else {
		return name, nil
	}
}

// generateServiceName removes any formatting from b and returns the first 4 bytes
// of its MD5 checksum.
func (sd *ServiceDefinition) generateServiceName() (string, error) {
	// remove formatting
	clean, err := json.Marshal(*sd)
	if err != nil {
		return "", mask(err)
	}

	// create hash
	s := md5.Sum(clean)
	return fmt.Sprintf("%x", s[0:4]), nil
}
