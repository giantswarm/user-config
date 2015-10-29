package userconfig

// ComponentDefinition represents either a runnable service inside a container or a
// component configuration
type ComponentDefinition struct {
	// Name of a docker image to use when running a container. The image includes
	// tags. E.g. dockerfile/redis:latest.
	Image *ImageDefinition `json:"image,omitempty" description:"Name of a docker image to use when running a container. The image includes tags."`

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
	Domains V2DomainDefinitions `json:"domains,omitempty" description:"List of domains to bind exposed ports to."`

	// Service names required by a service.
	Links LinkDefinitions `json:"links,omitempty" description:"List of dependencies of this service."`

	Expose ExposeDefinitions `json:"expose,omitempty" description:"List of port mappings to define a stable API."`

	Scale *ScaleDefinition `json:"scale,omitempty" description:"Scaling settings of the component."`

	Pod PodEnum `json:"pod,omitempty" description:"Pod behavior of this component and its children."`

	// If true, user needs to send a signal to indicate that the container is ready is should be considered running.
	SignalReady bool `json:"signal-ready,omitempty" description:"If true, user has to notify when the component is ready."`

	// How much memory to give the container. If empty, the server may decide on a default limit.
	MemoryLimit ByteSize `json:"memory-limit,omitempty" description:"How much memory to give the container. If empty, the server may decide on a default limit."`

	// NOTE: In case we add new fields to the component definition, we need to
	// implement proper diff functionality for those new fields as well.
}

// validate performs semantic validations of this ComponentDefinition.
// Return the first possible error.
func (nd *ComponentDefinition) validate(valCtx *ValidationContext) error {
	if nd.Image != nil {
		if err := nd.Image.Validate(valCtx); err != nil {
			return mask(err)
		}
	}

	if err := nd.validateMemoryLimit(valCtx); err != nil {
		return mask(err)
	}

	if err := nd.Ports.Validate(valCtx); err != nil {
		return mask(err)
	}

	if err := nd.Domains.validate(nd.Ports); err != nil {
		return mask(err)
	}

	if err := nd.Links.Validate(valCtx); err != nil {
		return mask(err)
	}

	if err := nd.Volumes.validate(valCtx); err != nil {
		return mask(err)
	}

	if nd.Scale != nil {
		if err := nd.Scale.validate(valCtx); err != nil {
			return mask(err)
		}
	}

	if err := nd.Expose.validate(); err != nil {
		return mask(err)
	}

	return nil
}

func (nd *ComponentDefinition) validateMemoryLimit(valCtx *ValidationContext) error {
	// An empty memory-limit is okay
	if nd.MemoryLimit.IsEmpty() {
		return nil
	}

	// Is the value itself valid?
	value, err := nd.MemoryLimit.Bytes()
	if err != nil {
		return mask(InvalidMemoryLimitError)
	}

	// If we have a validationContext, compare against boundaries
	if valCtx == nil {
		return nil
	}

	if !valCtx.EnableUserMemoryLimit {
		if !nd.MemoryLimit.IsEmpty() {
			return maskf(InvalidMemoryLimitError, "Providing a 'memory-limit' is not enabled.")
		}
		return nil
	}

	min, err := valCtx.MinMemoryLimit.Bytes()
	if err != nil {
		panic("Provided minimum memory-limit is invalid: " + err.Error())
	}
	max, err := valCtx.MaxMemoryLimit.Bytes()
	if err != nil {
		panic("Provided maximum memory-limit is invalid: " + err.Error())
	}

	if value < min {
		return maskf(InvalidMemoryLimitError, "memory-limit must be above %s", valCtx.MinMemoryLimit.String())
	}
	if value > max {
		return maskf(InvalidMemoryLimitError, "memory-limit must be below %s", valCtx.MaxMemoryLimit.String())
	}
	return nil
}

func (nd *ComponentDefinition) hideDefaults(valCtx *ValidationContext) *ComponentDefinition {
	if nd.Scale != nil {
		nd.Scale = nd.Scale.hideDefaults(valCtx)
	}

	return nd
}

func (nd *ComponentDefinition) setDefaults(valCtx *ValidationContext) {
	// set default scale definition if not set
	if nd.Scale == nil {
		nd.Scale = &ScaleDefinition{}
	}

	nd.Scale.setDefaults(valCtx)
}

// IsComponent returns true if the component has a defined container image, false otherwise.
func (nd *ComponentDefinition) IsComponent() bool {
	return nd.Image != nil
}

// IsPodRoot returns true if Pod is set to children or inherit.
func (nd *ComponentDefinition) IsPodRoot() bool {
	return nd.Pod == PodChildren || nd.Pod == PodInherit
}
