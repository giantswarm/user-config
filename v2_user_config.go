package userconfig

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/giantswarm/generic-types-go"
	"github.com/juju/errgo"
)

type V2AppDefinition struct {
	Includes []string        `json:"includes,omitempty"` // feature not implemented yet
	Nodes    map[string]Node `json:"nodes"`
}

// V2Service is a runnable service inside a container.
type V2Service struct {
	// Name of a docker image to use when running a container. The image includes
	// tags. E.g. dockerfile/redis:latest.
	Image generictypes.DockerImage `json:"image" description:"Name of a docker image to use when running a container. The image includes tags."`

	// If given, overwrite the entrypoint of the docker image.
	EntryPoint string `json:"entrypoint,omitempty" description:"If given, overwrite the entrypoint of the docker image."`

	// List of ports a service exposes. E.g. 6379/tcp
	Ports []generictypes.DockerPort `json:"ports,omitempty" description:"List of ports this service exposes."`

	// Docker env to inject into docker containers.
	Env EnvList `json:"env,omitempty" description:"List of environment variables used by this service."`

	// Docker volumes to inject into docker containers.
	Volumes []VolumeConfig `json:"volumes,omitempty" description:"List of volumes to attach to this service."`

	// Arguments for processes inside docker containers.
	Args []string `json:"args,omitempty" description:"List of arguments passed to the entry point of this service."`

	// Domains to bind the port to:  domainName => port, e.g. "www.heise.de" => "80"
	Domains DomainConfig `json:"domains,omitempty" description:"List of domains to bind exposed ports to."`

	// Service names required by a service.
	Links []DependencyConfig `json:"links,omitempty" description:"List of dependencies of this service."`
}

type ExposeConfig struct {
	Port     generictypes.DockerPort `json:"port" description:"Port of the stable API."`
	Node     string                  `json:"node" description:"Node name of the node that exposes a given port."`
	NodePort generictypes.DockerPort `json:"node_port" description:"Port of the given node."`
}

// Node is either a runnable service inside a container or a node configuration.
type Node struct {
	V2Service

	Expose []ExposeConfig       `json:"expose,omitempty" description:"List of port mappings to define a stable API."`
	Scale  *ScalingPolicyConfig `json:"scale,omitempty" description:"Scaling settings of the node"`
}

func ParseV2AppName(b []byte) (string, error) {
	s := md5.Sum(b)
	return fmt.Sprintf("%x", s[0:4]), nil
}

// ParseV2AppDefinition parses the v2 app definition.
func ParseV2AppDefinition(b []byte) (V2AppDefinition, error) {
	var appDef V2AppDefinition
	if err := json.Unmarshal(b, &appDef); err != nil {
		if IsSyntaxError(err) {
			if strings.Contains(err.Error(), "$") {
				return V2AppDefinition{}, errgo.WithCausef(nil, err, "Cannot parse swarm.json. Maybe not all variables replaced properly.")
			}
		}

		return V2AppDefinition{}, Mask(err)
	}

	return appDef, nil
}
