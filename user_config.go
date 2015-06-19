package userconfig

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	generictypes "github.com/giantswarm/generic-types-go" // Added `generictypes` alias here as a workaround for swagger in api docs
	"github.com/juju/errgo"
)

type AppDefinition struct {
	AppName     string            `json:"app_name" description:"Application name"`
	PublicPorts map[string]string `json:"public_ports,omitempty" description:"Port mappings"`
	Services    []ServiceConfig   `json:"services" description:"List of service that are part of this application"`
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

// ParseV1AppDefinition parses the v1 app definition.
func ParseV1AppDefinition(byteSlice []byte) (AppDefinition, error) {
	var app AppDefinition
	if err := json.Unmarshal(byteSlice, &app); err != nil {
		if IsSyntax(err) {
			if strings.Contains(err.Error(), "$") {
				return AppDefinition{}, errgo.WithCausef(nil, err, "Cannot parse swarm.json. Maybe not all variables replaced properly.")
			}
		}

		return AppDefinition{}, Mask(err)
	}

	return app, nil
}

// User defined service.
type ServiceConfig struct {
	ServiceName string            `json:"service_name" description:"Name of the service"`
	PublicPorts map[string]string `json:"public_ports,omitempty" description:"Port mappings"`

	// Config defining how many instances should be launched.
	ScalingPolicy *ScaleDefinition `json:"scaling_policy,omitempty" description:"Scaling settings of all components in this service"`

	Components []ComponentConfig `json:"components" description:"List of components that are part of this service"`
}

// FindComponent finds a component with given name if the list of components inside this service.
// it returns nil if not found
func (sc *ServiceConfig) FindComponent(name string) *ComponentConfig {
	return sc.findComponent(name)
}

type DependencyConfig struct {
	// Name of a required component
	Name string `json:"name" description:"Name of a required component"`

	// The name how this dependency should appear in the container
	Alias string `json:"alias,omitempty" description:"The name how this dependency should appear in the container"`

	// Port of the required component
	Port generictypes.DockerPort `json:"port" description:"Port of the required component"`

	// Wether the component should run on the same machine
	SameMachine bool `json:"same_machine,omitempty" description:"Wether the component should run on the same machine"`
}

func (ld DependencyConfig) validate() error {
	if ld.Name == "" {
		return Mask(errgo.WithCausef(nil, InvalidDependencyConfigError, "link name must not be empty"))
	}

	// for easy validation we create a port definitions type and use its
	// validate method
	pds := PortDefinitions{ld.Port}
	if err := pds.validate(); err != nil {
		return Mask(errgo.WithCausef(nil, InvalidDependencyConfigError, "invalid link: %s", err.Error()))
	}

	return nil
}

type ComponentConfig struct {
	// Name of a component.
	ComponentName string `json:"component_name" description:"Name of the component"`

	// Config defining how many instances should be launched.
	ScalingPolicy *ScaleDefinition `json:"scaling_policy,omitempty" description:"Scaling settings of the component"`

	InstanceConfig

	PodConfig
}

// GetAllMountPoints creates a list of all mount points of a component.
func (cc *ComponentConfig) GetAllMountPoints(service *ServiceConfig) ([]string, error) {
	visitedComponents := make(map[string]string)
	return cc.getAllMountPoints(service, visitedComponents)
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

	return errgo.WithCausef(err, InvalidEnvListFormatError, "")
}

type InstanceConfig struct {
	// Name of a docker image to use when running a container. The image includes
	// tags. E.g. dockerfile/redis:latest.
	Image generictypes.DockerImage `json:"image" description:"Name of a docker image to use when running a container. The image includes tags"`

	// If given, overwrite the entrypoint of the docker image.
	EntryPoint string `json:"entrypoint,omitempty" description:"If given, overwrite the entrypoint of the docker image"`

	// List of ports a service exposes. E.g. 6379/tcp
	Ports []generictypes.DockerPort `json:"ports,omitempty" description:"List of ports this component exposes"`

	// Docker env to inject into docker containers.
	Env EnvList `json:"env,omitempty" description:"List of environment variables used by this component"`

	// Docker volumes to inject into docker containers.
	Volumes []VolumeConfig `json:"volumes,omitempty" description:"List of volumes to attach to this component"`

	// Arguments for processes inside docker containers.
	Args []string `json:"args,omitempty" description:"List of arguments passed to the entry point of this component"`

	// Domains to bind the port to:  domainName => port, e.g. "www.heise.de" => "80"
	Domains DomainDefinitions `json:"domains,omitempty" description:"List of domains to bind exposed ports to"`

	// Service names required by a service.
	Dependencies []DependencyConfig `json:"dependencies,omitempty" description:"List of dependencies of this component"`
}

type PodConfig struct {
	// Name of the pod a component will join.
	// An empty name means that the component will not join any shared pod and create its own.
	PodName string `json:"pod,omitempty" description:"Name of the pod to join"`
}
