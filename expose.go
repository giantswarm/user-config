package userconfig

import (
	"github.com/giantswarm/generic-types-go"
)

type ExposeDefinition struct {
	Port     generictypes.DockerPort `json:"port" description:"Port of the stable API."`
	Node     NodeName                `json:"node,omitempty" description:"Name of the node that implements the stable API."`
	NodePort generictypes.DockerPort `json:"node_port,omitempty" description:"Port of the given node that implements the stable API."`
}

type ExposeDefinitions []ExposeDefinition

// validateExpose
func (nds NodeDefinitions) validateExpose() error {
	for nodeName, node := range nds {
		// detect invalid exposes
		for _, expose := range node.Expose {
			// Try to find the implementation node
			implName := expose.Node
			var implNode *NodeDefinition
			if implName.Empty() {
				// Expose refers to own node
				implNode = node
			} else {
				// Implementation node refers to a child node
				if !implName.IsChildOf(nodeName) {
					return maskf(InvalidNodeDefinitionError, "invalid expose to node '%s': is not a child of '%s'", implName, nodeName)
				}
				// Find implementation node
				var err error
				implNode, err = nds.NodeByName(implName)
				if err != nil {
					return maskf(InvalidNodeDefinitionError, "invalid expose to node '%s': does not exists", implName)
				}
			}

			// Does the implementation node expose the targeted port?
			implPort := expose.ImplementationPort()
			if !implNode.Ports.contains(implPort) {
				return maskf(InvalidNodeDefinitionError, "invalid expose to node '%s': does not export port '%s'", implName, implPort)
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
			return maskf(InvalidNodeDefinitionError, "cannot expose with empty port")
		}

		for j := i + 1; j < len(eds); j++ {
			if eds[j].Port.Equals(ed.Port) {
				// Duplicate exposed port found
				return maskf(InvalidNodeDefinitionError, "port %s is exposed more than once", ed.Port)
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

// ImplementationNodeName returns the name of the node that implements the stable API exposed by this definition.
func (ed *ExposeDefinition) ImplementationNodeName(containingNodeName NodeName) NodeName {
	if ed.Node.Empty() {
		return containingNodeName
	}
	return ed.Node
}

// ImplementationPort returns the port on the node that implements the stable API exposed by this definition.
func (ed *ExposeDefinition) ImplementationPort() generictypes.DockerPort {
	if ed.NodePort.Empty() {
		return ed.Port
	}
	return ed.NodePort
}
