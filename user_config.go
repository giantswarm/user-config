package userconfig

import (
	"encoding/json"
)

type AppConfig struct {
	AppName     string            `json:"app_name"`
	PublicPorts map[string]string `json:"public_ports,omitempty"`
	Services    []ServiceConfig   `json:"services"`
}

// We need to define a separate AppConfig type that does not implement the
// json.Unmarshaler. We do this because we need to call json.Unmarshal in the
// unmarshaler and would create a infinite loop using the original AppConfig
// type.
type AppConfigCopy AppConfig

func (ac *AppConfig) UnmarshalJSON(b []byte) error {
	b, err := FixJSONFieldNames(b)
	if err != nil {
		return Mask(err)
	}

	if err := CheckForUnknownFields(b, ac); err != nil {
		return Mask(err)
	}

	// Just unmarshal the given bytes into the app-config struct, since there
	// were no errors.
	var appConfigCopy AppConfigCopy
	if err := json.Unmarshal(b, &appConfigCopy); err != nil {
		return Mask(err)
	}

	*ac = AppConfig(appConfigCopy)

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

// We need to define a separate ServiceConfig type that does not implement the
// json.Unmarshaler. We do this because we need to call json.Unmarshal in the
// unmarshaler and would create a infinite loop using the original ServiceConfig
// type.
type ServiceConfigCopy ServiceConfig

func (sc *ServiceConfig) UnmarshalJSON(b []byte) error {
	b, err := FixJSONFieldNames(b)
	if err != nil {
		return Mask(err)
	}

	// Just unmarshal the given bytes into the service-config struct, since there
	// were no errors.
	var serviceConfigCopy ServiceConfigCopy
	if err := json.Unmarshal(b, &serviceConfigCopy); err != nil {
		return Mask(err)
	}

	*sc = ServiceConfig(serviceConfigCopy)

	return nil
}

type VolumeConfig struct {
	// Path of the volume to mount, e.g. "/opt/service/".
	Path string `json:"path"`

	// Storage size in GB, e.g. "5 GB".
	Size string `json:"size"`
}

type DependencyConfig struct {
	// Name of a required component
	Name string `json:"name"`

	// The name how this dependency should appear in the container
	Alias string `json:"alias,omitempty"`

	// Port of the required component
	Port int `json:"port"`

	// Wether the component should run on the same machine
	SameMachine bool `json:"same_machine,omitempty"`
}

type ComponentConfig struct {
	// Name of a service.
	ComponentName string `json:"component_name"`

	// Config defining how many instances should be launched.
	ScalingPolicy *ScalingPolicyConfig `json:"scaling_policy,omitempty"`

	InstanceConfig
}

// We need to define a separate ComponentConfig type that does not implement the
// json.Unmarshaler. We do this because we need to call json.Unmarshal in the
// unmarshaler and would create a infinite loop using the original ComponentConfig
// type.
type ComponentConfigCopy ComponentConfig

func (sc *ComponentConfig) UnmarshalJSON(b []byte) error {
	b, err := FixJSONFieldNames(b)
	if err != nil {
		return Mask(err)
	}

	// Just unmarshal the given bytes into the component-config struct, since there
	// were no errors.
	var componentConfigCopy ComponentConfigCopy
	if err := json.Unmarshal(b, &componentConfigCopy); err != nil {
		return Mask(err)
	}

	*sc = ComponentConfig(componentConfigCopy)

	return nil
}

type InstanceConfig struct {
	// Name of a docker image to use when running a container. The image includes
	// tags. E.g. dockerfile/redis:latest.
	Image DockerImage `json:"image"`

	// List of ports a service exposes. E.g. 6379/tcp
	Ports []string `json:"ports,omitempty"`

	// Docker env to inject into docker containers.
	Env []string `json:"env,omitempty"`

	// Docker volumes to inject into docker containers.
	Volumes []VolumeConfig `json:"volumes,omitempty"`

	// Arguments for processes inside docker containers.
	Args []string `json:"args,omitempty"`

	// Domains to bind the port to:  domainName => port, e.g. "www.heise.de" => "80"
	Domains map[string]string `json:"domains,omitempty"`

	// Service names required by a service.
	Dependencies []DependencyConfig `json:"dependencies,omitempty"`
}
