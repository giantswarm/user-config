package userconfig

import (
	"encoding/json"

	"github.com/giantswarm/generic-types-go"
	"github.com/juju/errgo"
)

type PortDefinitions []generictypes.DockerPort
type portDefinitions PortDefinitions

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

// UnmarshalJSON performs custom unmarshalling to support smart
// data types.
func (pds *PortDefinitions) UnmarshalJSON(data []byte) error {
	if data[0] != '[' {
		// Must be a single value, convert to an array of one
		newData := []byte{}
		newData = append(newData, '[')
		newData = append(newData, data...)
		newData = append(newData, ']')

		data = newData
	}

	var local portDefinitions
	if err := json.Unmarshal(data, &local); err != nil {
		return mask(err)
	}
	*pds = PortDefinitions(local)

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
		podNodes, err := nds.PodNodes(nodeName)
		if err != nil {
			return mask(err)
		}
		list := LinkDefinitions{}
		for _, pn := range podNodes {
			if pn.Links == nil {
				// No dependencies
				continue
			}
			list = append(list, pn.Links...)
		}

		// Check list for duplicates
		for i, l1 := range list {
			alias1, err := l1.LinkName()
			if err != nil {
				return mask(err)
			}
			for j := i + 1; j < len(list); j++ {
				l2 := list[j]
				alias2, err := l2.LinkName()
				if err != nil {
					return mask(err)
				}
				if alias1 == alias2 {
					// Same alias, Port must match and Name must match
					if !l1.TargetPort.Equals(l2.TargetPort) {
						return maskf(InvalidDependencyConfigError, "duplicate (with different ports) dependency '%s' in pod under '%s'", alias1, nodeName.String())
					}
					if l1.Node != l2.Node {
						return maskf(InvalidDependencyConfigError, "duplicate (with different names) dependency '%s' in pod under '%s'", alias1, nodeName.String())
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
		podNodes, err := nds.PodNodes(nodeName)
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
					return maskf(InvalidPortConfigError, "multiple nodes export port '%s' in pod under '%s'", port1.String(), nodeName.String())
				}
			}
		}
	}

	// No errors detected
	return nil
}
