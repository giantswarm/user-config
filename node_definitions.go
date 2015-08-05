package userconfig

type NodeDefinitions map[NodeName]*NodeDefinition

func (nds NodeDefinitions) validate(valCtx *ValidationContext) error {
	for nodeName, _ := range nds {
		if err := nodeName.Validate(); err != nil {
			return mask(err)
		}

		// because of defaulting when validating we need to reference the to the
		// address of the node. so its changes effect the app definition after
		// parsing.
		if err := nds[nodeName].validate(valCtx); err != nil {
			return mask(err)
		}
	}

	if err := nds.validateLinks(); err != nil {
		return mask(err)
	}

	if err := nds.validateExpose(); err != nil {
		return mask(err)
	}

	if err := nds.validateVolumesRefs(); err != nil {
		return mask(err)
	}

	if err := nds.validateUniqueMountPoints(); err != nil {
		return mask(err)
	}

	// Check for duplicate exposed ports in pods
	if err := nds.validateUniquePortsInPods(); err != nil {
		return mask(err)
	}

	// Check dependencies in pods
	if err := nds.validateUniqueDependenciesInPods(); err != nil {
		return mask(err)
	}

	// Check node relations in pods
	if err := nds.validatePods(); err != nil {
		return mask(err)
	}

	// Check scaling policies in pods
	if err := nds.validateScalingPolicyInPods(); err != nil {
		return mask(err)
	}

	// Check leafs
	if err := nds.validateLeafs(); err != nil {
		return mask(err)
	}

	return nil
}

func (nds NodeDefinitions) hideDefaults(valCtx *ValidationContext) NodeDefinitions {
	for nodeName, node := range nds {
		nds[nodeName] = node.hideDefaults(valCtx)
	}

	return nds
}

func (nds NodeDefinitions) setDefaults(valCtx *ValidationContext) {
	for nodeName, _ := range nds {
		nds[nodeName].setDefaults(valCtx)
	}
}

func (nds *NodeDefinitions) NodeByName(name NodeName) (*NodeDefinition, error) {
	for nodeName, nodeDef := range *nds {
		if name == nodeName {
			return nodeDef, nil
		}
	}

	return nil, maskf(NodeNotFoundError, name.String())
}

// ParentOf returns the closest parent of the node with the given name.
// If there is no such node, a NodeNotFoundError is returned.
func (nds *NodeDefinitions) ParentOf(name NodeName) (NodeName, *NodeDefinition, error) {
	for {
		parentName, err := name.ParentName()
		if err != nil {
			return "", nil, maskf(NodeNotFoundError, "'%s' has no parent", name)
		}
		if parent, err := nds.NodeByName(parentName); err == nil {
			return parentName, parent, nil
		}
		name = parentName
	}
	return "", nil, maskf(NodeNotFoundError, "'%s' has no parent", name)
}

// IsRoot returns true if the given node name has no more parent nodes
// in this set of nodes.
func (nds *NodeDefinitions) IsRoot(name NodeName) bool {
	_, _, err := nds.ParentOf(name)
	return err != nil
}

// FilterNodes returns a set of all my nodes for which the given predicate returns true.
func (nds *NodeDefinitions) FilterNodes(predicate func(nodeName NodeName, nodeDef NodeDefinition) bool) NodeDefinitions {
	list := make(NodeDefinitions)
	for nodeName, nodeDef := range *nds {
		if predicate(nodeName, *nodeDef) {
			list[nodeName] = nodeDef
		}
	}
	return list
}

// ChildNodes returns a map of all nodes that are a direct child of a node with
// the given name.
func (nds *NodeDefinitions) ChildNodes(name NodeName) NodeDefinitions {
	return nds.FilterNodes(func(nodeName NodeName, nodeDef NodeDefinition) bool {
		return nodeName.IsDirectChildOf(name)
	})
}

// ChildNodesRecursive returns a list of all nodes that are a direct child of a node with
// the given name and all child nodes of this children (recursive).
func (nds *NodeDefinitions) ChildNodesRecursive(name NodeName) NodeDefinitions {
	return nds.FilterNodes(func(nodeName NodeName, nodeDef NodeDefinition) bool {
		return nodeName.IsChildOf(name)
	})
}

// PodNodes returns a map of all nodes that are part of the pod specified by a node with
// the given name.
func (nds *NodeDefinitions) PodNodes(name NodeName) (NodeDefinitions, error) {
	parent, err := nds.NodeByName(name)
	if err != nil {
		return nil, mask(err)
	}
	switch parent.Pod {
	case PodChildren:
		// Collect all direct child nodes that do not have pod set to 'none'.
		return nds.FilterNodes(func(nodeName NodeName, nodeDef NodeDefinition) bool {
			return nodeName.IsDirectChildOf(name) && nodeDef.Pod != PodNone
		}), nil
	case PodInherit:
		// Collect all child nodes that do not have pod set to 'none'.
		noneNames := []NodeName{}
		children := nds.FilterNodes(func(nodeName NodeName, nodeDef NodeDefinition) bool {
			if !nodeName.IsChildOf(name) {
				return false
			}
			if nodeDef.Pod == PodNone {
				noneNames = append(noneNames, nodeName)
				return false
			}
			return true
		})
		// We now  go over the list and remove all children that have some parent with pod='none'
		for _, nodeName := range noneNames {
			for childName, _ := range children {
				if childName.IsChildOf(nodeName) {
					// Child of pod='none', remove from list
					delete(children, childName)
				}
			}
		}
		return children, nil
	default:
		return nil, maskf(InvalidArgumentError, "Node '%s' a has no pod setting", name)
	}
}

// PodRoot returns the node that defines the pod the node with given name is a part of.
// If there is no such node, NodeNotFoundError is returned.
func (nds *NodeDefinitions) PodRoot(name NodeName) (NodeName, *NodeDefinition, error) {
	for {
		// Find first parent
		parentName, parent, err := nds.ParentOf(name)
		if err != nil {
			return "", nil, err
		}
		if parent.IsPodRoot() {
			// We found our pod root
			return parentName, parent, nil
		}
		// Not a pood root, continue up the tree
		name = parentName
	}
}

// IsLeaf returns true if the node with the given name has no children,
// false otherwise.
func (nds *NodeDefinitions) IsLeaf(name NodeName) bool {
	for nodeName, _ := range *nds {
		if nodeName.IsChildOf(name) {
			return false
		}
	}
	return true
}

// MountPoints returns a list of all mount points of a node, that is given by
// name
func (nds *NodeDefinitions) MountPoints(name NodeName) ([]string, error) {
	visited := make(map[string]string)
	return nds.mountPointsRecursive(name, visited)
}

// mountPointsRecursive creates a list of all mount points of a node
func (nds *NodeDefinitions) mountPointsRecursive(name NodeName, visited map[string]string) ([]string, error) {
	// prevent cycles
	if _, ok := visited[name.String()]; ok {
		return nil, maskf(VolumeCycleError, "volume cycle detected in '%s'", name)
	}
	visited[name.String()] = name.String()

	node, err := nds.NodeByName(name)
	if err != nil {
		return nil, mask(err)
	}

	// get all mountpoints
	mountPoints := []string{}
	for _, vol := range node.Volumes {
		if vol.Path != "" {
			mountPoints = append(mountPoints, normalizeFolder(vol.Path))
		} else if vol.VolumePath != "" {
			mountPoints = append(mountPoints, normalizeFolder(vol.VolumePath))
		} else if vol.VolumesFrom != "" {
			p, err := nds.mountPointsRecursive(NodeName(vol.VolumesFrom), visited)
			if err != nil {
				return nil, err
			}
			mountPoints = append(mountPoints, p...)
		}
	}
	return mountPoints, nil
}
