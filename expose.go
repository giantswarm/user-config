package userconfig

import (
	"github.com/giantswarm/generic-types-go"
)

type ExposeDefinition struct {
	Port       generictypes.DockerPort `json:"port" description:"Port of the stable API."`
	Component  ComponentName           `json:"component,omitempty" description:"Name of the component that implements the stable API."`
	TargetPort generictypes.DockerPort `json:"target_port,omitempty" description:"Port of the given component that implements the stable API."`
}

type ExposeDefinitions []ExposeDefinition

// validateExpose
func (nds ComponentDefinitions) validateExpose() error {
	rootComponents := []*ComponentDefinition{}
	for componentName, component := range nds {
		// detect invalid exposes
		for _, expose := range component.Expose {
			// Try to find the implementation component
			implName := expose.Component
			var implComponent *ComponentDefinition
			if implName.Empty() {
				// Expose refers to own component
				implComponent = component
			} else {
				// Implementation component refers to a child component
				if !implName.IsChildOf(componentName) {
					return maskf(InvalidComponentDefinitionError, "invalid expose to component '%s': is not a child of '%s'", implName, componentName)
				}
				// Find implementation component
				var err error
				implComponent, err = nds.ComponentByName(implName)
				if err != nil {
					return maskf(InvalidComponentDefinitionError, "invalid expose to component '%s': does not exists", implName)
				}
			}

			// Does the implementation component expose the targeted port?
			implPort := expose.ImplementationPort()
			if !implComponent.Ports.contains(implPort) {
				return maskf(InvalidComponentDefinitionError, "invalid expose to component '%s': does not export port '%s'", implName, implPort)
			}
		}

		// Collect root components
		if nds.IsRoot(componentName) {
			rootComponents = append(rootComponents, component)
		}
	}

	// Check for duplicate exposed ports on root components
	for i, component := range rootComponents {
		for _, expose := range component.Expose {
			for j := i + 1; j < len(rootComponents); j++ {
				if rootComponents[j].Expose.contains(expose.Port) {
					return maskf(InvalidComponentDefinitionError, "port '%s' is exposed by multiple root components", expose.Port)
				}
			}
		}
	}

	return nil
}

// validate checks for invalid and duplicate entries
func (eds ExposeDefinitions) validate() error {
	for i, ed := range eds {
		if ed.Port.Empty() {
			// Invalid exposed port found
			return maskf(InvalidComponentDefinitionError, "cannot expose with empty port")
		}

		for j := i + 1; j < len(eds); j++ {
			if eds[j].Port.Equals(ed.Port) {
				// Duplicate exposed port found
				return maskf(InvalidComponentDefinitionError, "port '%s' is exposed more than once", ed.Port)
			}
		}
	}
	return nil
}

// contains returns true if the given list of expose definitions contains
// a definition that exposes the given port.
func (eds ExposeDefinitions) contains(port generictypes.DockerPort) bool {
	for _, ed := range eds {
		if ed.Port.Equals(port) {
			return true
		}
	}

	return false
}

// defByPort returns the first expose definition in this list that equals the given port.
// If no such definition is found, a PortNotFoundError is returned.
func (eds ExposeDefinitions) defByPort(port generictypes.DockerPort) (ExposeDefinition, error) {
	for _, ed := range eds {
		if ed.Port.Equals(port) {
			return ed, nil
		}
	}

	return ExposeDefinition{}, maskf(PortNotFoundError, "port '%s' not found", port)
}

// ImplementationComponentName returns the name of the component that implements the stable API exposed by this definition.
func (ed *ExposeDefinition) ImplementationComponentName(containingComponentName ComponentName) ComponentName {
	if ed.Component.Empty() {
		return containingComponentName
	}
	return ed.Component
}

// ImplementationPort returns the port on the component that implements the stable API exposed by this definition.
func (ed *ExposeDefinition) ImplementationPort() generictypes.DockerPort {
	if ed.TargetPort.Empty() {
		return ed.Port
	}
	return ed.TargetPort
}

// resolve resolves the implementation of the given Expose definition in the context of the given
// component definitions.
// Resolve returns the name of the component the implements this expose and its implementation port.
// If this expose definition cannot be resolved, an error is returned.
func (ed *ExposeDefinition) Resolve(containingComponentName ComponentName, nds ComponentDefinitions) (ComponentName, generictypes.DockerPort, error) {
	// Get implementation component name
	implName := ed.ImplementationComponentName(containingComponentName)
	// Get implementation port
	implPort := ed.ImplementationPort()

	// Find implementation component
	component, err := nds.ComponentByName(implName)
	if err != nil {
		return "", generictypes.DockerPort{}, mask(err)
	}

	// Check expose definitions of component
	if implExpDef, err := component.Expose.defByPort(implPort); err == nil {
		// Recurse into the implementation component
		return implExpDef.Resolve(implName, nds)
	}

	// Check exported ports of component
	if component.Ports.contains(implPort) {
		// Found implementation component and port
		return implName, implPort, nil
	}

	// Port is not exposed, not exported by implementation component
	return "", generictypes.DockerPort{}, maskf(PortNotFoundError, "component '%s' does not export port '%s'", implName, implPort)
}
