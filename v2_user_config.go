package userconfig

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"
)

type V2AppDefinition struct {
	// Optional app name
	AppName AppName `json:"name,omitempty"`

	// Components
	Components ComponentDefinitions `json:"components"`
}

func (ad *V2AppDefinition) UnmarshalJSON(data []byte) error {
	// We fix the json buffer so V2CheckForUnknownFields doesn't complain about
	// `Components` (with uper N).
	data, err := FixJSONFieldNames(data)
	if err != nil {
		return err
	}

	if err := V2CheckForUnknownFields(data, ad); err != nil {
		return err
	}

	// Just unmarshal the given bytes into the app def struct, since there
	// were no errors.
	var adc v2AppDefCopy
	if err := json.Unmarshal(data, &adc); err != nil {
		return mask(err)
	}

	result := V2AppDefinition(adc)

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

// validate performs semantic validations of this V2AppDefinition.
// Return the first possible error.
func (ad *V2AppDefinition) Validate(valCtx *ValidationContext) error {
	if len(ad.Components) == 0 {
		return maskf(InvalidAppDefinitionError, "components must not be empty")
	}

	if !ad.AppName.Empty() {
		if err := ad.AppName.Validate(); err != nil {
			return mask(err)
		}
	}

	if err := ad.Components.validate(valCtx); err != nil {
		return mask(err)
	}

	return nil
}

// HideDefaults uses the given validation context to determine what definition
// details should be hidden. The caller can clean the definition that way to
// not confuse the user with information he has not set by himself.
func (ad *V2AppDefinition) HideDefaults(valCtx *ValidationContext) (*V2AppDefinition, error) {
	if valCtx == nil {
		return &V2AppDefinition{}, maskf(MissingValidationContextError, "cannot hide defaults")
	}

	ad.Components = ad.Components.hideDefaults(valCtx)
	return ad, nil
}

// SetDefaults sets necessary default values if not given by the user.
func (ad *V2AppDefinition) SetDefaults(valCtx *ValidationContext) error {
	if valCtx == nil {
		return maskf(MissingValidationContextError, "cannot set defaults")
	}

	ad.Components.setDefaults(valCtx)
	return nil
}

// V2AppName returns the name of the given definition if it exists.
// It is does not exist, it generates an app name.
func V2AppName(b []byte) (string, error) {
	// parse and validate
	appDef, err := ParseV2AppDefinition(b)
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
// It is does not exist, it generates an application name.
func (ad *V2AppDefinition) Name() (string, error) {
	// Is a name specified?
	if !ad.AppName.Empty() {
		return ad.AppName.String(), nil
	}

	// No name is specified, generate one
	if name, err := ad.generateAppName(); err != nil {
		return "", mask(err)
	} else {
		return name, nil
	}
}

// generateAppName removes any formatting from b and returns the first 4 bytes
// of its MD5 checksum.
func (ad *V2AppDefinition) generateAppName() (string, error) {
	// remove formatting
	clean, err := json.Marshal(*ad)
	if err != nil {
		return "", mask(err)
	}

	// create hash
	s := md5.Sum(clean)
	return fmt.Sprintf("%x", s[0:4]), nil
}

// ParseV2AppDefinition tries to parse the v2 app definition.
func ParseV2AppDefinition(b []byte) (V2AppDefinition, error) {
	var appDef V2AppDefinition
	if err := json.Unmarshal(b, &appDef); err != nil {
		if IsSyntax(err) {
			if strings.Contains(err.Error(), "$") {
				return V2AppDefinition{}, maskf(err, "Cannot parse swarm.json. Maybe not all variables replaced properly.")
			}
		}

		return V2AppDefinition{}, mask(err)
	}

	return appDef, nil
}
