package userconfig

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/giantswarm/generic-types-go"
)

type V2AppDefinition struct {
	Nodes NodeDefinitions `json:"nodes"`
}

func (ad *V2AppDefinition) UnmarshalJSON(data []byte) error {
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
	if len(ad.Nodes) == 0 {
		return maskf(InvalidAppDefinitionError, "nodes must not be empty")
	}

	if err := ad.Nodes.validate(valCtx); err != nil {
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

	ad.Nodes = ad.Nodes.hideDefaults(valCtx)
	return ad, nil
}

// SetDefaults sets necessary default values if not given by the user.
func (ad *V2AppDefinition) SetDefaults(valCtx *ValidationContext) error {
	if valCtx == nil {
		return maskf(MissingValidationContextError, "cannot set defaults")
	}

	ad.Nodes.setDefaults(valCtx)
	return nil
}

type ExposeDefinition struct {
	Port     generictypes.DockerPort `json:"port" description:"Port of the stable API."`
	Node     string                  `json:"node" description:"Node name of the node that exposes a given port."`
	NodePort generictypes.DockerPort `json:"node_port" description:"Port of the given node."`
}

// V2GenerateAppName removes any formatting from b and returns the first 4 bytes
// of its MD5 checksum.
func V2GenerateAppName(b []byte) (string, error) {
	// parse and validate
	appDef, err := ParseV2AppDefinition(b)
	if err != nil {
		return "", mask(err)
	}

	// remove formatting
	clean, err := json.Marshal(appDef)
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
