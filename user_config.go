package userconfig

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/giantswarm/docker-types-go"
	"github.com/juju/errgo"
)

type AppDefinition struct {
	AppName     string            `json:"app_name"`
	PublicPorts map[string]string `json:"public_ports,omitempty"`
	Services    []ServiceConfig   `json:"services"`
}

func (ac *AppDefinition) UnmarshalJSON(data []byte) error {
	// We fix the json buffer so CheckForUnknownFields doesn't complain about `Components`.
	data, err := FixJSONFieldNames(data)
	if err != nil {
		return err
	}

	if err := CheckForUnknownFields(data, ac); err != nil {
		return err
	}

	// Just unmarshal the given bytes into the app-config struct, since there
	// were no errors.
	var appConfigCopy appConfigCopy
	if err := json.Unmarshal(data, &appConfigCopy); err != nil {
		return Mask(err)
	}

	result := AppDefinition(appConfigCopy)

	// Perform semantic checks
	if err := result.validate(); err != nil {
		return Mask(err)
	}

	*ac = result

	return nil
}

type ScalingPolicyConfig struct {
	// Minimum instances to launch.
	Min int `json:"min,omitempty"`

	// Maximum instances to launch.
	Max int `json:"max,omitempty"`
}

// User defined service.
type ServiceConfig struct {
	ServiceName string            `json:"service_name"`
	PublicPorts map[string]string `json:"public_ports,omitempty"`

	// Config defining how many instances should be launched.
	ScalingPolicy *ScalingPolicyConfig `json:"scaling_policy,omitempty"`

	Components []ComponentConfig `json:"components"`
}

type VolumeConfig struct {
	// Path of the volume to mount, e.g. "/opt/service/".
	Path string `json:"path"`

	// Storage size in GB, e.g. "5 GB".
	Size VolumeSize `json:"size"`
}

type DependencyConfig struct {
	// Name of a required component
	Name string `json:"name"`

	// The name how this dependency should appear in the container
	Alias string `json:"alias,omitempty"`

	// Port of the required component
	Port dockertypes.DockerPort `json:"port"`

	// Wether the component should run on the same machine
	SameMachine bool `json:"same_machine,omitempty"`
}

type DomainConfig map[string]dockertypes.DockerPort

type ComponentConfig struct {
	// Name of a service.
	ComponentName string `json:"component_name"`

	// Config defining how many instances should be launched.
	ScalingPolicy *ScalingPolicyConfig `json:"scaling_policy,omitempty"`

	InstanceConfig
}

// List of environment settings like "KEY=VALUE", "KEY2=VALUE2"
type EnvList []string

// UnmarshalJSON supports parsing an EnvList as array and as structure
func (this *EnvList) UnmarshalJSON(data []byte) error {
	var err error
	// Try to parse as struct first
	if len(data) > 1 && data[0] == '{' {
		kvMap := make(map[string]string)
		err = json.Unmarshal(data, &kvMap)
		if err == nil {
			// Success, wrap into array
			// Sort the keys first so the outcome it always the same
			keys := []string{}
			for k, _ := range kvMap {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			list := []string{}
			for _, k := range keys {
				v := kvMap[k]
				list = append(list, fmt.Sprintf("%s=%s", k, v))
			}
			*this = list
			return nil
		}
	}

	// Try to parse are []string
	if len(data) > 1 && data[0] == '[' {
		list := []string{}
		err = json.Unmarshal(data, &list)
		if err != nil {
			return err
		}
		*this = list
		return nil
	}

	return errgo.WithCausef(err, ErrInvalidEnvListFormat)
}

type InstanceConfig struct {
	// Name of a docker image to use when running a container. The image includes
	// tags. E.g. dockerfile/redis:latest.
	Image dockertypes.DockerImage `json:"image"`

	// List of ports a service exposes. E.g. 6379/tcp
	Ports []dockertypes.DockerPort `json:"ports,omitempty"`

	// Docker env to inject into docker containers.
	Env EnvList `json:"env,omitempty"`

	// Docker volumes to inject into docker containers.
	Volumes []VolumeConfig `json:"volumes,omitempty"`

	// Arguments for processes inside docker containers.
	Args []string `json:"args,omitempty"`

	// Domains to bind the port to:  domainName => port, e.g. "www.heise.de" => "80"
	Domains DomainConfig `json:"domains,omitempty"`

	// Service names required by a service.
	Dependencies []DependencyConfig `json:"dependencies,omitempty"`
}
