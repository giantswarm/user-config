package userconfig

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/giantswarm/generic-types-go"
	"github.com/juju/errgo"
)

var (
	nodeNameRegExp = regexp.MustCompile("^[a-z0-9A-Z_/-]{0,99}$")
)

type V2AppDefinition struct {
	Nodes NodeDefinitions `json:"nodes"`
}

func (ad *V2AppDefinition) UnmarshalJSON(data []byte) error {
	// We fix the json buffer so CheckForUnknownFields doesn't complain about `Components`.
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
		return Mask(err)
	}

	result := V2AppDefinition(adc)

	// Perform semantic checks
	if err := result.Validate(); err != nil {
		return Mask(err)
	}

	*ad = result

	return nil
}

// validate performs semantic validations of this V2AppDefinition.
// Return the first possible error.
func (ad *V2AppDefinition) Validate() error {
	if len(ad.Nodes) == 0 {
		return Mask(errgo.WithCausef(nil, InvalidAppDefinitionError, "nodes must not be empty"))
	}

	if err := ad.Nodes.validate(); err != nil {
		return Mask(err)
	}

	return nil
}

type NodeDefinitions map[NodeName]*NodeDefinition

func (nds NodeDefinitions) validate() error {
	for nodeName, node := range nds {
		if err := nodeName.validate(); err != nil {
			return Mask(err)
		}

		// because of defaulting when validating we need to reference the to the
		// address of the node. so its changes effect the app definition after
		// parsing.
		if err := nds[nodeName].validate(); err != nil {
			return err
		}

		// detect invalid links
		for _, link := range node.Links {
			nodeFound := false

			for nn, n := range nds {
				if link.Name == nn.String() {
					nodeFound = true

					if !n.Ports.contains(link.Port) {
						return Mask(errgo.WithCausef(nil, InvalidNodeDefinitionError, "invalid link to node '%s': does not export port '%s'", nodeName, link.Port))
					}
				}
			}

			if !nodeFound {
				return Mask(errgo.WithCausef(nil, InvalidNodeDefinitionError, "invalid link to node '%s': does not exists", link.Name))
			}
		}
	}

	return nil
}

type NodeName string

func (nn NodeName) String() string {
	return string(nn)
}

func (nn NodeName) validate() error {
	if nn == "" {
		return Mask(errgo.WithCausef(nil, InvalidNodeDefinitionError, "name must not be empty"))
	}

	if !nodeNameRegExp.MatchString(nn.String()) {
		return Mask(errgo.WithCausef(nil, InvalidNodeDefinitionError, "name '%s' must match regexp: %s", nn, nodeNameRegExp))
	}

	return nil
}

// Node is either a runnable service inside a container or a node definition.
type NodeDefinition struct {
	// Name of a docker image to use when running a container. The image includes
	// tags. E.g. dockerfile/redis:latest.
	Image generictypes.DockerImage `json:"image" description:"Name of a docker image to use when running a container. The image includes tags."`

	// If given, overwrite the entrypoint of the docker image.
	EntryPoint string `json:"entrypoint,omitempty" description:"If given, overwrite the entrypoint of the docker image."`

	// List of ports a service exposes. E.g. 6379/tcp
	Ports PortDefinitions `json:"ports,omitempty" description:"List of ports this service exposes."`

	// Docker env to inject into docker containers.
	Env EnvList `json:"env,omitempty" description:"List of environment variables used by this service."`

	// Docker volumes to inject into docker containers.
	Volumes VolumeDefinitions `json:"volumes,omitempty" description:"List of volumes to attach to this service."`

	// Arguments for processes inside docker containers.
	Args []string `json:"args,omitempty" description:"List of arguments passed to the entry point of this service."`

	// Domains to bind the port to:  domainName => port, e.g. "www.heise.de" => "80"
	Domains DomainDefinitions `json:"domains,omitempty" description:"List of domains to bind exposed ports to."`

	// Service names required by a service.
	Links LinkDefinitions `json:"links,omitempty" description:"List of dependencies of this service."`

	Expose []ExposeDefinition `json:"expose,omitempty" description:"List of port mappings to define a stable API."`

	Scale *ScaleDefinition `json:"scale,omitempty" description:"Scaling settings of the node"`
}

// validate performs semantic validations of this NodeDefinition.
// Return the first possible error.
func (nd *NodeDefinition) validate() error {
	if err := nd.Ports.validate(); err != nil {
		return Mask(err)
	}

	if err := nd.Domains.validate(nd.Ports); err != nil {
		return Mask(err)
	}

	if err := nd.Links.validate(); err != nil {
		return Mask(err)
	}

	if err := nd.Volumes.validate(); err != nil {
		return Mask(err)
	}

	// default scale definition if not set
	if nd.Scale == nil {
		nd.Scale = &ScaleDefinition{
			Min: MinScaleSize,
			Max: MaxScaleSize,
		}
	}

	if err := nd.Scale.validate(); err != nil {
		return Mask(err)
	}

	return nil
}

type ExposeDefinition struct {
	Port     generictypes.DockerPort `json:"port" description:"Port of the stable API."`
	Node     string                  `json:"node" description:"Node name of the node that exposes a given port."`
	NodePort generictypes.DockerPort `json:"node_port" description:"Port of the given node."`
}

// TODO Node.IsService() bool

// ParseV2AppName removes any formatting from b and returns the first 4 bytes
// of its MD5 checksum.
func ParseV2AppName(b []byte) (string, error) {
	appDef, err := ParseV2AppDefinition(b)
	if err != nil {
		return "", Mask(err)
	}

	clean, err := json.Marshal(appDef)
	if err != nil {
		return "", Mask(err)
	}

	s := md5.Sum(clean)
	return fmt.Sprintf("%x", s[0:4]), nil
}

// ParseV2AppDefinition tries to parse the v2 app definition.
func ParseV2AppDefinition(b []byte) (V2AppDefinition, error) {
	var appDef V2AppDefinition
	if err := json.Unmarshal(b, &appDef); err != nil {
		if IsSyntax(err) {
			if strings.Contains(err.Error(), "$") {
				return V2AppDefinition{}, errgo.WithCausef(nil, err, "Cannot parse swarm.json. Maybe not all variables replaced properly.")
			}
		}

		return V2AppDefinition{}, Mask(err)
	}

	return appDef, nil
}
