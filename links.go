package userconfig

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

			// Do the target node expose the linked to port?
			if !targetNode.Ports.contains(link.Port) {
				return maskf(InvalidNodeDefinitionError, "invalid link to node '%s': does not export port '%s'", link.Name, link.Port)
			}

			// Is the node allowed to link to the target node?
			if !isParentOrSiblingRecursive(nodeName, targetName) {
				return maskf(InvalidLinkDefinitionError, "invalid link to node '%s': node '%s' is not allowed to link to it", link.Name, nodeName)
			}
		}
	}
	return nil
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
