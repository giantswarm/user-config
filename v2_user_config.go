package userconfig

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"
)

type V2ServiceDefinition struct {
	// Optional service name
	ServiceName ServiceName `json:"name,omitempty"`

	// Nodes
	Nodes NodeDefinitions `json:"nodes"`
}

func (ad *V2ServiceDefinition) UnmarshalJSON(data []byte) error {
	// We fix the json buffer so V2CheckForUnknownFields doesn't complain about
	// `Nodes` (with uper N).
	data, err := FixJSONFieldNames(data)
	if err != nil {
		return err
	}

	if err := V2CheckForUnknownFields(data, ad); err != nil {
		return err
	}

	// Just unmarshal the given bytes into the app def struct, since there
	// were no errors.
	var adc v2ServiceDefCopy
	if err := json.Unmarshal(data, &adc); err != nil {
		return mask(err)
	}

	result := V2ServiceDefinition(adc)

	// validate app definition without validation context. validation context is
	// given on server side to additionally validate specific definitions.
	if err := result.Validate(nil); err != nil {
		return mask(err)
	}

	*ad = result

	return nil
}

type ValidationContext struct {
	Org       string
	Protocols []string

	MinScaleSize int
	MaxScaleSize int

	MinVolumeSize VolumeSize
	MaxVolumeSize VolumeSize

	PublicDockerRegistry  string
	PrivateDockerRegistry string
}

// validate performs semantic validations of this V2ServiceDefinition.
// Return the first possible error.
func (ad *V2ServiceDefinition) Validate(valCtx *ValidationContext) error {
	if len(ad.Nodes) == 0 {
		return maskf(InvalidAppDefinitionError, "nodes must not be empty")
	}

	if !ad.ServiceName.Empty() {
		if err := ad.ServiceName.Validate(); err != nil {
			return mask(err)
		}
	}

	if err := ad.Nodes.validate(valCtx); err != nil {
		return mask(err)
	}

	return nil
}

// HideDefaults uses the given validation context to determine what definition
// details should be hidden. The caller can clean the definition that way to
// not confuse the user with information he has not set by himself.
func (ad *V2ServiceDefinition) HideDefaults(valCtx *ValidationContext) (*V2ServiceDefinition, error) {
	if valCtx == nil {
		return &V2ServiceDefinition{}, maskf(MissingValidationContextError, "cannot hide defaults")
	}

	ad.Nodes = ad.Nodes.hideDefaults(valCtx)
	return ad, nil
}

// SetDefaults sets necessary default values if not given by the user.
func (ad *V2ServiceDefinition) SetDefaults(valCtx *ValidationContext) error {
	if valCtx == nil {
		return maskf(MissingValidationContextError, "cannot set defaults")
	}

	ad.Nodes.setDefaults(valCtx)
	return nil
}

// V2ServiceName returns the name of the given definition if it exists.
// It is does not exist, it generates an app name.
func V2ServiceName(b []byte) (string, error) {
	// parse and validate
	appDef, err := ParseV2ServiceDefinition(b)
	if err != nil {
		return "", mask(err)
	}

	// Get name
	if name, err := appDef.Name(); err != nil {
		return "", mask(err)
	} else {
		return name, nil
	}
}

// Name returns the name of the given definition if it exists.
// It is does not exist, it generates an service name.
func (ad *V2ServiceDefinition) Name() (string, error) {
	// Is a name specified?
	if !ad.ServiceName.Empty() {
		return ad.ServiceName.String(), nil
	}

	// No name is specified, generate one
	if name, err := ad.generateServiceName(); err != nil {
		return "", mask(err)
	} else {
		return name, nil
	}
}

// generateServiceName removes any formatting from b and returns the first 4 bytes
// of its MD5 checksum.
func (ad *V2ServiceDefinition) generateServiceName() (string, error) {
	// remove formatting
	clean, err := json.Marshal(*ad)
	if err != nil {
		return "", mask(err)
	}

	// create hash
	s := md5.Sum(clean)
	return fmt.Sprintf("%x", s[0:4]), nil
}

// ParseV2ServiceDefinition tries to parse the v2 app definition.
func ParseV2ServiceDefinition(b []byte) (V2ServiceDefinition, error) {
	var appDef V2ServiceDefinition
	if err := json.Unmarshal(b, &appDef); err != nil {
		if IsSyntax(err) {
			if strings.Contains(err.Error(), "$") {
				return V2ServiceDefinition{}, maskf(err, "Cannot parse swarm.json. Maybe not all variables replaced properly.")
			}
		}

		return V2ServiceDefinition{}, mask(err)
	}

	return appDef, nil
}
