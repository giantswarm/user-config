package userconfig

import (
	"github.com/giantswarm/generic-types-go"
)

type LinkDefinitions []DependencyConfig

func (lds LinkDefinitions) Validate(valCtx *ValidationContext) error {
	links := map[string]bool{}

	for _, link := range lds {
		if err := link.Validate(valCtx); err != nil {
			return mask(err)
		}

		// detect duplicated link name
		name := link.Alias
		if name == "" {
			name = link.Name
		}
		if _, ok := links[name]; ok {
			return maskf(InvalidLinkDefinitionError, "duplicated link: %s", name)
		}
		links[name] = true
	}

	return nil
}

// Resolve resolves the implementation of the given link in the context of the given
// node definitions.
// Resolve returns the name of the node that implements this link and its implementation port.
// If this link cannot be resolved, an error is returned.
func (link DependencyConfig) Resolve(nds NodeDefinitions) (NodeName, generictypes.DockerPort, error) {
	// Resolve initial link target
	targetName := NodeName(link.Name)
	targetNode, err := nds.NodeByName(targetName)
	if err != nil {
		return "", generictypes.DockerPort{}, maskf(NodeNotFoundError, link.Name)
	}

	// If the linked to port exposed by the target node?
	if expDef, err := targetNode.Expose.defByPort(link.Port); err == nil {
		// Link to exposed port, let expose definition resolve this further
		return expDef.resolve(targetName, nds)
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
