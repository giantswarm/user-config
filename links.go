package userconfig

import (
	"github.com/giantswarm/generic-types-go"
)

type LinkDefinition struct {
	Service ServiceName `json:"service,omitempty" description:"Name of the service that is linked to"`

	// Name of a required node
	Name NodeName `json:"name,omitempty" description:"Name of a node that is linked to"`

	// The name how this dependency should appear in the container
	Alias string `json:"alias,omitempty" description:"The name how this dependency should appear in the container"`

	// Port of the required node
	Port generictypes.DockerPort `json:"port" description:"Port of the required node"`
}

type LinkDefinitions []LinkDefinition

func (ld LinkDefinition) Validate(valCtx *ValidationContext) error {
	if ld.Name.Empty() && ld.Service.Empty() {
		return maskf(InvalidLinkDefinitionError, "link name must not be empty")
	}
	if !ld.Name.Empty() {
		if err := ld.Name.Validate(); err != nil {
			return maskf(InvalidLinkDefinitionError, "invalid link name: %s", err.Error())
		}
	}
	if !ld.Service.Empty() {
		if err := ld.Service.Validate(); err != nil {
			return maskf(InvalidLinkDefinitionError, "invalid link service: %s", err.Error())
		}
	}
	if !ld.Name.Empty() && !ld.Service.Empty() {
		return maskf(InvalidLinkDefinitionError, "link service and name cannot be set both")
	}

	// for easy validation we create a port definitions type and use its
	// validate method
	pds := PortDefinitions{ld.Port}
	if err := pds.Validate(valCtx); err != nil {
		return maskf(InvalidLinkDefinitionError, "invalid link: %s", err.Error())
	}

	return nil
}

// LinkName returns the name of this link as it will be used inside
// the node.
// This defaults to the alias. If that is not specified, the local name
// of the Node name will be used, or if that is also empty, the app name.
func (ld LinkDefinition) LinkName() (string, error) {
	if ld.Alias != "" {
		return ld.Alias, nil
	}
	if !ld.Name.Empty() {
		// Take the dependency name from the last part of the node name
		// (using `LocalName()`).
		// This is done to prevent that the dependency name has '/' in it.
		return ld.Name.LocalName().String(), nil
	}
	if !ld.Service.Empty() {
		return ld.Service.String(), nil
	}
	return "", mask(InvalidLinkDefinitionError)
}

// LinksToOtherService returns true if this definition defines
// a link between a node an another service.
func (ld LinkDefinition) LinksToOtherService() bool {
	return !ld.Service.Empty()
}

// LinksToSameService returns true if this definition defines
// a link between a node and another node within the same service.
func (ld LinkDefinition) LinksToSameService() bool {
	return ld.Service.Empty()
}

func (lds LinkDefinitions) Validate(valCtx *ValidationContext) error {
	links := map[string]bool{}

	for _, link := range lds {
		if err := link.Validate(valCtx); err != nil {
			return mask(err)
		}

		// detect duplicated link name
		linkName, err := link.LinkName()
		if err != nil {
			return mask(err)
		}
		if _, ok := links[linkName]; ok {
			return maskf(InvalidLinkDefinitionError, "duplicated link: %s", linkName)
		}
		links[linkName] = true
	}

	return nil
}

// Resolve resolves the implementation of the given link in the context of the given
// node definitions.
// Resolve returns the name of the node that implements this link and its implementation port.
// If this link cannot be resolved, an error is returned.
func (link LinkDefinition) Resolve(nds NodeDefinitions) (NodeName, generictypes.DockerPort, error) {
	// Resolve initial link target
	targetName := link.Name
	targetNode, err := nds.NodeByName(targetName)
	if err != nil {
		return "", generictypes.DockerPort{}, maskf(NodeNotFoundError, link.Name.String())
	}

	// If the linked to port exposed by the target node?
	if expDef, err := targetNode.Expose.defByPort(link.Port); err == nil {
		// Link to exposed port, let expose definition resolve this further
		return expDef.Resolve(targetName, nds)
	}

	if targetNode.Ports.contains(link.Port) {
		// Link points directly to an exported port of the target
		return targetName, link.Port, nil
	}

	// Invalid link
	return "", generictypes.DockerPort{}, maskf(InvalidLinkDefinitionError, "port %s not found in %s", link.Port, targetName)
}

// validateLinks
func (nds NodeDefinitions) validateLinks() error {
	for nodeName, node := range nds {
		// detect invalid links
		for _, link := range node.Links {
			// If the link is inter-service, we cannot validate it here.
			if link.LinksToOtherService() {
				continue
			}

			// Try to find the target node
			targetName := NodeName(link.Name)
			targetNode, err := nds.NodeByName(targetName)
			if err != nil {
				return maskf(InvalidNodeDefinitionError, "invalid link to node '%s': does not exists", link.Name)
			}

			// Does the target node expose the linked to port?
			if !targetNode.Expose.contains(link.Port) && !targetNode.Ports.contains(link.Port) {
				return maskf(InvalidNodeDefinitionError, "invalid link to node '%s': does not export port '%s'", link.Name, link.Port)
			}

			// Is the node allowed to link to the target node?
			if !isLinkAllowed(nodeName, targetName) {
				return maskf(InvalidLinkDefinitionError, "invalid link to node '%s': node '%s' is not allowed to link to it", link.Name, nodeName)
			}
		}
	}
	return nil
}

// isLinkAllowed returns true if a node with given name is allowed to
// link to a node with given target name.
func isLinkAllowed(nodeName, targetName NodeName) bool {
	// If target is a child or grand child of node, it is ok.
	if targetName.IsChildOf(nodeName) {
		return true
	}

	// If target is a parent/sibling ("up or right-left"), it is ok.
	if isParentOrSiblingRecursive(nodeName, targetName) {
		return true
	}

	return false
}

// isParentOrSiblingRecursive returns true if targetName is a parent of nodeName,
// or targetName is a sibling of node name.
// The test is done recursively.
func isParentOrSiblingRecursive(nodeName, targetName NodeName) bool {
	if nodeName.IsSiblingOf(targetName) {
		return true
	}
	parentName, err := nodeName.ParentName()
	if err != nil {
		// No more parent
		return false
	}
	return isParentOrSiblingRecursive(parentName, targetName)
}
