package userconfig

import (
	"github.com/giantswarm/generic-types-go"
	"github.com/juju/errgo"
)

type PortDefinitions []generictypes.DockerPort

// Validate tries to validate the current PortDefinitions. If valCtx is nil,
// nothing can be validated. The given valCtx must provide at least one
// protocol, or Validate returns an error. The currently valid one should only
// be TCP.
func (pds PortDefinitions) Validate(valCtx *ValidationContext) error {
	if valCtx == nil {
		return nil
	}

	if len(valCtx.Protocols) == 0 {
		return errgo.Newf("missing protocol in validation context")
	}

	for _, port := range pds {
		if !contains(valCtx.Protocols, port.Protocol) {
			return maskf(InvalidPortConfigError, "invalid protocol '%s' for port '%s', expected one of %v", port.Protocol, port.Port, valCtx.Protocols)
		}
	}

	return nil
}

func (pds PortDefinitions) contains(port generictypes.DockerPort) bool {
	for _, pd := range pds {
		// generictypes.DockerPort implements Equals to properly compare the
		// format "<port>/<protocol>"
		if pd.Equals(port) {
			return true
		}
	}

	return false
}

func contains(protocols []string, protocol string) bool {
	for _, p := range protocols {
		if p == protocol {
			return true
		}
	}

	return false
}

// validateUniqueDependenciesInPods checks that there are no dependencies with same alias and different port/name
func (nds *NodeDefinitions) validateUniqueDependenciesInPods() error {
	for nodeName, nodeDef := range *nds {
		if !nodeDef.IsPodRoot() {
			continue
		}

		// Collect all dependencies in this pod
		podNodes, err := nds.PodNodes(nodeName.String())
		if err != nil {
			return mask(err)
		}
		list := []DependencyConfig{}
		for _, pn := range podNodes {
			if pn.Links == nil {
				// No dependencies
				continue
			}
			list = append(list, pn.Links...)
		}

		// Check list for duplicates
		for i, dep1 := range list {
			alias1 := dep1.alias(nodeName.String())
			for j := i + 1; j < len(list); j++ {
				dep2 := list[j]
				alias2 := dep2.alias(nodeName.String())
				if alias1 == alias2 {
					// Same alias, Port must match and Name must match
					if !dep1.Port.Equals(dep2.Port) {
						return maskf(InvalidDependencyConfigError, "Cannot parse app config. Duplicate (but different ports) dependency '%s' in pod under '%s'.", alias1, nodeName.String())
					}
					if dep1.Name != dep2.Name {
						return maskf(InvalidDependencyConfigError, "Cannot parse app config. Duplicate (but different names) dependency '%s' in pod under '%s'.", alias1, nodeName.String())
					}
				}
			}
		}
	}

	// No errors detected
	return nil
}

// validateUniquePortsInPods checks that there are no duplicate ports in a single pod
func (nds *NodeDefinitions) validateUniquePortsInPods() error {
	for nodeName, nodeDef := range *nds {
		if !nodeDef.IsPodRoot() {
			continue
		}

		// Collect all ports in this pod
		podNodes, err := nds.PodNodes(nodeName.String())
		if err != nil {
			return mask(err)
		}
		list := []generictypes.DockerPort{}
		for _, pn := range podNodes {
			if pn.Ports == nil {
				// No dependencies
				continue
			}
			list = append(list, pn.Ports...)
		}

		// Check list for duplicates
		for i, port1 := range list {
			for j := i + 1; j < len(list); j++ {
				port2 := list[j]
				if port1.Equals(port2) {
					return maskf(InvalidPortConfigError, "Cannot parse app config. Multiple nodes export port '%s' in pod under '%s'.", port1.String(), nodeName.String())
				}
			}
		}
	}

	// No errors detected
	return nil
}
