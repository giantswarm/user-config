package userconfig

type AppConfig struct {
	AppName     string            `json:"app_name"`
	PublicPorts map[string]string `json:"public_ports"`
	Services    []ServiceConfig   `json:"services"`
}

type ScalingPolicyConfig struct {
	// Minimum instances to launch.
	Min int

	// Maximum instances to launch.
	Max int
}

// User defined service.
type ServiceConfig struct {
	ServiceName string            `json:"service_name"`
	PublicPorts map[string]string `json:"public_ports"`

	// Config defining how many instances should be launched.
	ScalingPolicy ScalingPolicyConfig `json:"scaling_policy"`

	Components []ComponentConfig
}

type VolumeConfig struct {
	// Path of the volume to mount, e.g. "/opt/service/".
	Path string

	// Storage size in GB, e.g. "5 GB".
	Size string
}

type DependencyConfig struct {
	// Name of a required component
	Name string `json:"name"`

	// Port of the required component
	Port int `json:"port"`

	// Wether the component should run on the same machine
	SameMachine bool `json:"same_machine,omitempty"`
}

type ComponentConfig struct {
	// Name of a service.
	ComponentName string `json:"component_name"`

	// Name of a docker image to use when running a container. The image includes
	// tags. E.g. dockerfile/redis:latest.
	Image string `json:"image"`

	// Config defining how many instances should be launched.
	ScalingPolicy ScalingPolicyConfig `json:"scaling_policy"`

	// List of ports a service exposes. E.g. 6379/tcp
	Ports []string `json:"ports"`

	// Docker env to inject into docker containers.
	Env []string `json:"env"`

	// Docker volumes to inject into docker containers.
	Volumes []VolumeConfig `json:"volumes"`

	// Arguments for processes inside docker containers.
	Args []string `json:"args"`

	// Domains to bind the port to:  domainName => port, e.g. "www.heise.de" => "80"
	Domains map[string]string `json:"domains,omitempty"`

	// Service names required by a service.
	Dependencies []DependencyConfig `json:"dependencies,omitempty"`
}
