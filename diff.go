package userconfig

import (
	"reflect"
)

type DiffType string

const (
	// Used to indicate that the app-name in the two files was changed
	InfoAppNameChanged DiffType = "infoAppNameChanged"

	// Used to indicate that a component (service or component) was added. `Name` in the DiffInfo describes the component that was added.
	InfoComponentAdded DiffType = "component-added"

	// Used to indicate that a component (service or component) was removed. `Name` in the DiffInfo describes the component that was removed.
	InfoComponentRemoved DiffType = "component-removed"

	// Used to indicate that the scaling config of a component changed.
	InfoComponentScalingUpdated DiffType = "component-scaling-changed"

	// Used to indicate that the InstanceConfig in the ComponentConfig identified by `Name` changed.
	InfoInstanceConfigUpdated DiffType = "instance-update"

	// Used to indicate that the component itself changed (but not the underlying InstanceConfig). `Name` in the DiffInfo describes the component that was changed.
	InfoComponentUpdated DiffType = "component-update"
)

type DiffInfo struct {
	// What type changed: app, service, component
	Type DiffType

	// Path to the component that changed.
	Name []string
}

// Diff compares the two AppConfigs and returns a list of changes between the two.
func Diff(newConfig, oldConfig AppDefinition) []DiffInfo {
	if newConfig.AppName != oldConfig.AppName {
		return []DiffInfo{
			DiffInfo{Type: InfoAppNameChanged, Name: []string{oldConfig.AppName}},
		}
	}

	changes := []DiffInfo{}
	c := make(chan DiffInfo)
	go func() {
		diffHierarchy([]string{newConfig.AppName}, appComponent(newConfig), appComponent(oldConfig), c)
		close(c)
	}()

	for change := range c {
		changes = append(changes, change)
	}
	return changes
}

type component interface {
	Name() string
	Children() []component

	// Diff is called by diffHierarchy when two components have the same name and should be checked
	// for internal differences. Diffs should be written to changes.
	Diff(path []string, other component, changes chan<- DiffInfo)
}

type appComponent AppDefinition

func (n appComponent) Name() string { return n.AppName }
func (n appComponent) Children() []component {
	components := []component{}
	for _, service := range n.Services {
		components = append(components, serviceComponent(service))
	}
	return components
}
func (n appComponent) Diff(path []string, other component, changes chan<- DiffInfo) {
	diffHierarchy(path, n, other, changes)
}

type serviceComponent ServiceConfig

func (s serviceComponent) Name() string { return s.ServiceName }
func (s serviceComponent) Children() []component {
	components := []component{}
	for _, component := range s.Components {
		components = append(components, componentComponent(component))
	}
	return components
}
func (n serviceComponent) Diff(path []string, other component, changes chan<- DiffInfo) {
	diffHierarchy(path, n, other, changes)
}

type componentComponent ComponentConfig

func (c componentComponent) Name() string { return c.ComponentName }
func (c componentComponent) Children() []component {
	components := []component{}
	return components
}

// Diff compares c with other. If the InstanceConfigs differ, a DiffInfo of type
// infoInstanceConfigUpdated will be send to DiffInfo.
// If the InstanceConfig is the equal but the components itself differ, an infoComponentUpdated
// is sent.
// This allows to differ between changes that need to be applied to existing instances
// and infos that relates to scaling (at the moment).
func (c componentComponent) Diff(path []string, other component, changes chan<- DiffInfo) {
	otherComponent := other.(componentComponent)

	instanceConfigChanged := !reflect.DeepEqual(c.InstanceConfig, otherComponent.InstanceConfig)
	componentScalingChanged := !reflect.DeepEqual(c.ScalingPolicy, otherComponent.ScalingPolicy)
	componentChanged := !reflect.DeepEqual(c, other)

	if instanceConfigChanged {
		changes <- DiffInfo{Type: InfoInstanceConfigUpdated, Name: path}
	}
	if componentScalingChanged {
		changes <- DiffInfo{Type: InfoComponentScalingUpdated, Name: path}
	}

	// NOTE: This shouldn't trigger at the moment, a component only consist of scalingpolicy + instanceconfig
	// This is just here if we extend the component in the future so it also gets reported
	if !instanceConfigChanged && !componentScalingChanged && componentChanged {
		changes <- DiffInfo{Type: InfoComponentUpdated, Name: path}
	}
}

func diffHierarchy(path []string, newComponent, oldComponent component, changes chan<- DiffInfo) {
	oldComponents := map[string]component{}
	for _, child := range oldComponent.Children() {
		oldComponents[child.Name()] = child
	}

	// Search for components that were added or changed
	for _, child := range newComponent.Children() {
		name := child.Name()

		oldComponent, oldComponentExists := oldComponents[name]
		if oldComponentExists {
			child.Diff(append(path, name), oldComponent, changes)

			// This helps us later to determine which components were removed
			delete(oldComponents, name)
		} else {
			changes <- DiffInfo{Type: InfoComponentAdded, Name: append(path, name)}
		}
	}

	// Catch all components that were removed (the once left do not longer exist in the new config)
	for name, _ := range oldComponents {
		changes <- DiffInfo{Type: InfoComponentRemoved, Name: append(path, name)}
	}
}
