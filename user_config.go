package userconfig

type AppConfig struct {
	AppName     string                 `json:"app_name" bson:"app_name"`
	PublicPorts map[string]string      `json:"public_ports,omitempty" bson:"public_ports,omitempty"`
	Services    []ServiceConfig        `json:"services" bson:"services"`
	Arbitrary   map[string]interface{} `json:"-" bson:",inline"`
}

func (ac *AppConfig) UnmarshalJSON(b []byte) error {
	if err := UnmarshalWithBSONUnmarshaler(b, ac); err != nil {
		return Mask(err)
	}

	if err := CheckForUnknownFields(ac); err != nil {
		return Mask(err)
	}

	return nil
}

type ScalingPolicyConfig struct {
	// Minimum instances to launch.
	Min int `json:"min,omitempty" bson:"min,omitempty"`

	// Maximum instances to launch.
	Max int `json:"max,omitempty" bson:"max,omitempty"`
}

// User defined service.
type ServiceConfig struct {
	ServiceName string            `json:"service_name" bson:"service_name"`
	PublicPorts map[string]string `json:"public_ports,omitempty" bson:"public_ports,omitempty"`

	// Config defining how many instances should be launched.
	ScalingPolicy *ScalingPolicyConfig `json:"scaling_policy,omitempty" bson:"scaling_policy,omitempty"`

	Components []ComponentConfig `json:"components" bson:"components"`
}

type VolumeConfig struct {
	// Path of the volume to mount, e.g. "/opt/service/".
	Path string `json:"path" bson:"path"`

	// Storage size in GB, e.g. "5 GB".
	Size string `json:"size" bson:"size"`
}

type DependencyConfig struct {
	// Name of a required component
	Name string `json:"name" bson:"name"`

	// The name how this dependency should appear in the container
	Alias string `json:"alias,omitempty" bson:"alias,omitempty"`

	// Port of the required component
	Port int `json:"port" bson:"port"`

	// Wether the component should run on the same machine
	SameMachine bool `json:"same_machine,omitempty" bson:"same_machine,omitempty"`
}

type ComponentConfig struct {
	// Name of a service.
	ComponentName string `json:"component_name" bson:"component_name"`

	// Config defining how many instances should be launched.
	ScalingPolicy *ScalingPolicyConfig `json:"scaling_policy,omitempty" bson:"scaling_policy,omitempty"`

	// Name of a docker image to use when running a container. The image includes
	// tags. E.g. dockerfile/redis:latest.
	Image DockerImage `json:"image" bson:"image"`

	// List of ports a service exposes. E.g. 6379/tcp
	Ports []string `json:"ports,omitempty" bson:"ports,omitempty"`

	// Docker env to inject into docker containers.
	Env []string `json:"env,omitempty" bson:"env,omitempty"`

	// Docker volumes to inject into docker containers.
	Volumes []VolumeConfig `json:"volumes,omitempty" bson:"volumes,omitempty"`

	// Arguments for processes inside docker containers.
	Args []string `json:"args,omitempty" bson:"args,omitempty"`

	// Domains to bind the port to:  domainName => port, e.g. "www.heise.de" => "80"
	Domains map[string]string `json:"domains,omitempty" bson:"domains,omitempty"`

	// Service names required by a service.
	Dependencies []DependencyConfig `json:"dependencies,omitempty" bson:"dependencies,omitempty"`
}
