package userconfig

import (
	"github.com/juju/errgo"
)

type ComponentDefinitions map[ComponentName]*ComponentDefinition

func (nds ComponentDefinitions) validate(valCtx *ValidationContext) error {
	for componentName, _ := range nds {
		if err := componentName.Validate(); err != nil {
			return mask(err)
		}

		// because of defaulting when validating we need to reference the to the
		// address of the component. so its changes effect the app definition after
		// parsing.
		if err := nds[componentName].validate(valCtx); err != nil {
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

	// Check component relations in pods
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

// hideDefaults goes over each component and removes default values.
func (nds ComponentDefinitions) hideDefaults(valCtx *ValidationContext) ComponentDefinitions {
	for componentName, component := range nds {
		nds[componentName] = component.hideDefaults(valCtx)
	}

	return nds
}

// setDefaults goes over each component and applies default values
// if no values are set.
func (nds ComponentDefinitions) setDefaults(valCtx *ValidationContext) {
	// Share defaults in pods
	for componentName, component := range nds {
		if component.IsPodRoot() {
			// Found the start of a pod, make sure we share explicitly set values all over the pod
			podComponents, err := nds.PodComponents(componentName)
			if err == nil {
				nds.shareExplicitSettingsInPod(podComponents, valCtx)
			}
			// If err is not nil, we don't handle it here because we did not want to change the API of this
			// function. Not sharing values here will cause validation errors later, so we can safely
			// leave it out here.
			// The validation errors will be "scaling values cannot be different for components in a pod"
		}
	}

	// Apply all other defaults
	nds.doSetDefaults(valCtx)
}

// shareExplicitSettingsInPod goes over each component in a given pod and
// shares all explicitly set values in it,
func (nds ComponentDefinitions) shareExplicitSettingsInPod(podComponents ComponentDefinitions, valCtx *ValidationContext) {
	// Find explicitly set values
	localValCtx := *valCtx
	hasExplicitValues := false
	for _, c := range podComponents {
		if c.Scale != nil {
			if c.Scale.Min != 0 {
				localValCtx.MinScaleSize = c.Scale.Min
				hasExplicitValues = true
			}
			if c.Scale.Max != 0 {
				localValCtx.MaxScaleSize = c.Scale.Max
				hasExplicitValues = true
			}
			if c.Scale.Placement != "" {
				localValCtx.Placement = c.Scale.Placement
				hasExplicitValues = true
			}
		}
	}
	// Are there are any explicit values?
	if hasExplicitValues {
		podComponents.doSetDefaults(&localValCtx)
	}
}

// doSetDefaults goes over each component and applies default values
// if no values are set.
func (nds ComponentDefinitions) doSetDefaults(valCtx *ValidationContext) {
	for componentName, _ := range nds {
		nds[componentName].setDefaults(valCtx)
	}
}

func (nds *ComponentDefinitions) ComponentByName(name ComponentName) (*ComponentDefinition, error) {
	for componentName, componentDef := range *nds {
		if name == componentName {
			return componentDef, nil
		}
	}

	return nil, maskf(ComponentNotFoundError, name.String())
}

func (nds *ComponentDefinitions) Contains(name ComponentName) bool {
	for componentName, _ := range *nds {
		if name == componentName {
			return true
		}
	}

	return false
}

// ParentOf returns the closest parent of the component with the given name.
// If there is no such component, a ComponentNotFoundError is returned.
func (nds *ComponentDefinitions) ParentOf(name ComponentName) (ComponentName, *ComponentDefinition, error) {
	for {
		parentName, err := name.ParentName()
		if err != nil {
			return "", nil, maskf(ComponentNotFoundError, "'%s' has no parent", name)
		}
		if parent, err := nds.ComponentByName(parentName); err == nil {
			return parentName, parent, nil
		}
		name = parentName
	}
	return "", nil, maskf(ComponentNotFoundError, "'%s' has no parent", name)
}

// IsRoot returns true if the given component name has no more parent components
// in this set of components.
func (nds *ComponentDefinitions) IsRoot(name ComponentName) bool {
	_, _, err := nds.ParentOf(name)
	return err != nil
}

// FilterComponents returns a set of all my components for which the given predicate returns true.
func (nds *ComponentDefinitions) FilterComponents(predicate func(componentName ComponentName, componentDef ComponentDefinition) bool) ComponentDefinitions {
	list := make(ComponentDefinitions)
	for componentName, componentDef := range *nds {
		if predicate(componentName, *componentDef) {
			list[componentName] = componentDef
		}
	}
	return list
}

// ChildComponents returns a map of all components that are a direct child of a component with
// the given name.
func (nds *ComponentDefinitions) ChildComponents(name ComponentName) ComponentDefinitions {
	return nds.FilterComponents(func(componentName ComponentName, componentDef ComponentDefinition) bool {
		return componentName.IsDirectChildOf(name)
	})
}

// ChildComponentsRecursive returns a list of all components that are a direct child of a component with
// the given name and all child components of this children (recursive).
func (nds *ComponentDefinitions) ChildComponentsRecursive(name ComponentName) ComponentDefinitions {
	return nds.FilterComponents(func(componentName ComponentName, componentDef ComponentDefinition) bool {
		return componentName.IsChildOf(name)
	})
}

// PodComponents returns a map of all components that are part of the pod specified by a component with
// the given name.
func (nds *ComponentDefinitions) PodComponents(name ComponentName) (ComponentDefinitions, error) {
	parent, err := nds.ComponentByName(name)
	if err != nil {
		return nil, mask(err)
	}
	switch parent.Pod {
	case PodChildren:
		// Collect all direct child components that do not have pod set to 'none'.
		return nds.FilterComponents(func(componentName ComponentName, componentDef ComponentDefinition) bool {
			return componentName.IsDirectChildOf(name) && componentDef.Pod != PodNone
		}), nil
	case PodInherit:
		// Collect all child components that do not have pod set to 'none'.
		noneNames := []ComponentName{}
		children := nds.FilterComponents(func(componentName ComponentName, componentDef ComponentDefinition) bool {
			if !componentName.IsChildOf(name) {
				return false
			}
			if componentDef.Pod == PodNone {
				noneNames = append(noneNames, componentName)
				return false
			}
			return true
		})
		// We now  go over the list and remove all children that have some parent with pod='none'
		for _, componentName := range noneNames {
			for childName, _ := range children {
				if childName.IsChildOf(componentName) {
					// Child of pod='none', remove from list
					delete(children, childName)
				}
			}
		}
		return children, nil
	default:
		return nil, maskf(InvalidArgumentError, "Component '%s' a has no pod setting", name)
	}
}

// PodComponentsRecursive returns a map of all components that are part of the
// pod specified by a component with the given name. Other than
// ComponentDefinitions.PodComponents, this method does at first reverse lookup
// the pod root.
func (nds *ComponentDefinitions) PodComponentsRecursive(name ComponentName) (ComponentDefinitions, error) {
	rootName, _, err := nds.PodRoot(name)
	if err != nil {
		return nil, maskAny(err)
	}
	podComps, err := nds.PodComponents(rootName)
	if err != nil {
		return nil, maskAny(err)
	}
	return podComps, nil
}

// IsPartOfPod returns true in case the given component is part of a pod, otherwise
// false.
func (nds *ComponentDefinitions) IsPartOfPod(name ComponentName) bool {
	_, _, err := nds.PodRoot(name)
	if IsComponentNotFound(err) {
		return false
	}

	return true
}

// PodRoot returns the component that defines the pod the component with given name is a part of.
// If there is no such component, ComponentNotFoundError is returned.
func (nds *ComponentDefinitions) PodRoot(name ComponentName) (ComponentName, *ComponentDefinition, error) {
	for {
		// Find first parent
		parentName, parent, err := nds.ParentOf(name)
		if err != nil {
			return "", nil, maskAny(err)
		}
		if parent.IsPodRoot() {
			// We found our pod root
			return parentName, parent, nil
		}
		// Not a pood root, continue up the tree
		name = parentName
	}
}

// IsLeaf returns true if the component with the given name has no children,
// false otherwise.
func (nds *ComponentDefinitions) IsLeaf(name ComponentName) bool {
	for componentName, _ := range *nds {
		if componentName.IsChildOf(name) {
			return false
		}
	}
	return true
}

// MountPoints returns a list of all mount points of a component, that is given by
// name
func (nds *ComponentDefinitions) MountPoints(name ComponentName) ([]string, error) {
	visited := make(map[string]string)
	return nds.mountPointsRecursive(name, visited)
}

// mountPointsRecursive creates a list of all mount points of a component
func (nds *ComponentDefinitions) mountPointsRecursive(name ComponentName, visited map[string]string) ([]string, error) {
	// prevent cycles
	if _, ok := visited[name.String()]; ok {
		return nil, maskf(VolumeCycleError, "volume cycle detected in '%s'", name)
	}
	visited[name.String()] = name.String()

	component, err := nds.ComponentByName(name)
	if err != nil {
		return nil, mask(err)
	}

	// get all mountpoints
	mountPoints := []string{}
	for _, vol := range component.Volumes {
		if vol.Path != "" {
			mountPoints = append(mountPoints, normalizeFolder(vol.Path))
		} else if vol.VolumePath != "" {
			mountPoints = append(mountPoints, normalizeFolder(vol.VolumePath))
		} else if vol.VolumesFrom != "" {
			p, err := nds.mountPointsRecursive(ComponentName(vol.VolumesFrom), visited)
			if err != nil {
				return nil, err
			}
			mountPoints = append(mountPoints, p...)
		}
	}
	return mountPoints, nil
}

func (nds *ComponentDefinitions) ComponentNames() ComponentNames {
	compNames := ComponentNames{}

	for name, _ := range *nds {
		compNames = append(compNames, name)
	}

	return compNames
}

// AllDefsPerPod tries to group component definitions, based on this rules:
//   - group component definitions that share same pod
//   - prevent duplicated lists, once a component definition is present in one
//     list, it is not present in other lists.
// The resulting maps are sorted such that components that link to other components are
// found after the component that they link to.
func (nds *ComponentDefinitions) AllDefsPerPod(names ComponentNames) ([]ComponentDefinitions, error) {
	defsPerPod := []ComponentDefinitions{}

first:
	for _, name := range names {
		for _, defs := range defsPerPod {
			if defs.ComponentNames().Contain(name) {
				// if the current component is already tracked, skip it
				continue first
			}
		}

		if nds.IsPartOfPod(name) {
			podCompDefs, err := nds.PodComponentsRecursive(name)
			if err != nil {
				return nil, maskAny(err)
			}
			defsPerPod = append(defsPerPod, podCompDefs)
		} else {
			// The component definition for the given name does not define a pod.
			// Just group the current definiton on its own.
			compDef, err := nds.ComponentByName(name)
			if err != nil {
				return nil, maskAny(err)
			}
			defsPerPod = append(defsPerPod, ComponentDefinitions{name: compDef})
		}
	}

	sortedDefsPerPod, err := nds.sortByLinks(defsPerPod)
	if err != nil {
		return nil, maskAny(err)
	}
	return sortedDefsPerPod, nil
}

// sortByLinks orders the given list such that components with links to other components
// come later that the components they link to
func (nds *ComponentDefinitions) sortByLinks(defs []ComponentDefinitions) ([]ComponentDefinitions, error) {
	// Re-order such that components that link to other components are after before those other components
	for i := 0; i < len(defs); {
		def := defs[i]
		newIndex, err := nds.getIndexFromLinks(def, defs)
		if err != nil {
			return nil, maskAny(err)
		}
		if newIndex < 0 || newIndex <= i {
			// No change needed
			i++
			continue
		}
		// Swap defs[i] with defs[newIndex]
		defs[newIndex], defs[i] = defs[i], defs[newIndex]
		// Restart from the beginning
		i = 0
	}
	return defs, nil
}

// getIndexFromLinks returns the index (in cs) where the given component should be
// placed in order to be after all of the component it links to
func (nds *ComponentDefinitions) getIndexFromLinks(def ComponentDefinitions, defs []ComponentDefinitions) (int, error) {
	newIndex := -1

	// indexOf returns the index in `deps` of the map that contains a component with given name
	indexOf := func(compName ComponentName) (int, error) {
		for i, d := range defs {
			if d.Contains(compName) {
				return i, nil
			}
		}
		return 0, maskAny(ComponentNotFoundError)
	}

	for _, c := range def {
		for _, link := range c.Links {
			if link.LinksToOtherService() {
				continue
			}
			// Resolve link to link in same application
			implName, _, err := link.Resolve(*nds)
			if err != nil {
				return 0, maskAny(err)
			}

			// Find implementation component
			implDefIndex, err := indexOf(implName)
			if err != nil {
				// linked-to component is not in our lists, so we don't have to care about it
				continue
				//return 0, maskAny(errgo.WithCausef(nil, ComponentNotFoundError, "unknown component: %s in definitions list", implName))
			}

			if implDefIndex > newIndex {
				newIndex = implDefIndex
			}
		}
	}
	return newIndex, nil
}

func (nds *ComponentDefinitions) Map(names ComponentNames) ComponentDefinitions {
	list := ComponentDefinitions{}

	for name, def := range *nds {
		if names.Contain(name) {
			list[name] = def
		}
	}

	return list
}
