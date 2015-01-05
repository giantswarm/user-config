package userconfig

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/juju/errgo"
)

var (
	volumeSizeRegex = regexp.MustCompile("([0-9]+)\\s*(GB|G)")
)

type AppConfig struct {
	AppName     string            `json:"app_name"`
	PublicPorts map[string]string `json:"public_ports,omitempty"`
	Services    []ServiceConfig   `json:"services"`
}

func (ac *AppConfig) UnmarshalJSON(data []byte) error {
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

type VolumeSize string
type SizeUnit string

const (
	GB = SizeUnit("GB")
)

// UnmarshalJSON performs a format friendly parsing of volume sizes
func (this *VolumeSize) UnmarshalJSON(data []byte) error {
	var sz string
	err := json.Unmarshal(data, &sz)
	if err != nil {
		return err
	}
	sz = strings.ToUpper(sz)
	matches := volumeSizeRegex.FindStringSubmatch(sz)
	if matches == nil || len(matches) != 3 {
		return errgo.WithCausef(nil, ErrInvalidSize, "Cannot parse app config. Invalid size '%s' detected.", sz)
	}
	unit := matches[2]
	if unit == "G" {
		unit = "GB"
	}
	*this = VolumeSize(matches[1] + " " + unit)
	return nil
}

// Size gets the size part of a volume size as an integer.
// E.g. "5 GB" -> 5
func (this VolumeSize) Size() (int, error) {
	parts := strings.Split(string(this), " ")
	return strconv.Atoi(parts[0])
}

// Size gets the unit part of a volume size.
// E.g. "5 GB" -> GB
func (this VolumeSize) Unit() (SizeUnit, error) {
	parts := strings.Split(string(this), " ")
	if len(parts) < 2 {
		return GB, errgo.Newf("No unit found, got '%s'", string(this))
	}
	switch parts[1] {
	case "G":
		return GB, nil
	case "GB":
		return GB, nil
	default:
		return GB, errgo.Newf("Unknown unit, got '%s'", parts[1])
	}
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

// List of environment settings like "KEY=VALUE", "KEY2=VALUE2"
type EnvList []string

// UnmarshalJSON supports parsing an EnvList as array and as structure
func (this *EnvList) UnmarshalJSON(data []byte) error {
	// Try to parse as struct first
	if len(data) > 1 && data[0] == '{' {
		kvMap := make(map[string]string)
		err := json.Unmarshal(data, &kvMap)
		if err == nil {
			// Success, wrap into array
			list := []string{}
			for k, v := range kvMap {
				list = append(list, fmt.Sprintf("%s=%s", k, v))
			}
			*this = list
			return nil
		}
	}

	// Try to parse are []string
	if len(data) > 1 && data[0] == '[' {
		list := []string{}
		err := json.Unmarshal(data, &list)
		if err != nil {
			return err
		}
		*this = list
		return nil
	}

	return ErrInvalidFormat
}

type InstanceConfig struct {
	// Name of a docker image to use when running a container. The image includes
	// tags. E.g. dockerfile/redis:latest.
	Image DockerImage `json:"image"`

	// List of ports a service exposes. E.g. 6379/tcp
	Ports []string `json:"ports,omitempty"`

	// Docker env to inject into docker containers.
	Env EnvList `json:"env,omitempty"`

	// Docker volumes to inject into docker containers.
	Volumes []VolumeConfig `json:"volumes,omitempty"`

	// Arguments for processes inside docker containers.
	Args []string `json:"args,omitempty"`

	// Domains to bind the port to:  domainName => port, e.g. "www.heise.de" => "80"
	Domains map[string]string `json:"domains,omitempty"`

	// Service names required by a service.
	Dependencies []DependencyConfig `json:"dependencies,omitempty"`
}
