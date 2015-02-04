package userconfig

import (
	"reflect"
)

type DiffType string

const (
	// Used to indicate that the app-name in the two files was changed
	InfoAppNameChanged DiffType = "infoAppNameChanged"

	// Used to indicate that a node (service or component) was added. `Name` in the DiffInfo describes the node that was added.
	InfoNodeAdded DiffType = "node-added"

	// Used to indicate that a node (service or component) was removed. `Name` in the DiffInfo describes the node that was removed.
	InfoNodeRemoved DiffType = "node-removed"

	// Used to indicate that the scaling config of a component changed.
	InfoComponentScalingUpdated DiffType = "component-scaling-changed"

	// Used to indicate that the InstanceConfig in the ComponentConfig identified by `Name` changed.
	InfoInstanceConfigUpdated DiffType = "instance-update"

	// Used to indicate that the component itself changed (but not the underlying InstanceConfig). `Name` in the DiffInfo describes the node that was changed.
	InfoComponentUpdated DiffType = "component-update"
)

type DiffInfo struct {
	// What type changed: app, service, component
	Type DiffType

	// Path to the node that changed.
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
		diffHierarchy([]string{newConfig.AppName}, appNode(newConfig), appNode(oldConfig), c)
		close(c)
	}()

	for change := range c {
		changes = append(changes, change)
	}
	return changes
}

type node interface {
	Name() string
	Children() []node

	// Diff is called by diffHierarchy when two nodes have the same name and should be checked
	// for internal differences. Diffs should be written to changes.
	Diff(path []string, other node, changes chan<- DiffInfo)
}

type appNode AppDefinition

func (n appNode) Name() string { return n.AppName }
func (n appNode) Children() []node {
	nodes := []node{}
	for _, service := range n.Services {
		nodes = append(nodes, serviceNode(service))
	}
	return nodes
}
func (n appNode) Diff(path []string, other node, changes chan<- DiffInfo) {
	diffHierarchy(path, n, other, changes)
}

type serviceNode ServiceConfig

func (s serviceNode) Name() string { return s.ServiceName }
func (s serviceNode) Children() []node {
	nodes := []node{}
	for _, component := range s.Components {
		nodes = append(nodes, componentNode(component))
	}
	return nodes
}
func (n serviceNode) Diff(path []string, other node, changes chan<- DiffInfo) {
	diffHierarchy(path, n, other, changes)
}

type componentNode ComponentConfig

func (c componentNode) Name() string { return c.ComponentName }
func (c componentNode) Children() []node {
	nodes := []node{}
	return nodes
}

// Diff compares c with other. If the InstanceConfigs differ, a DiffInfo of type
// infoInstanceConfigUpdated will be send to DiffInfo.
// If the InstanceConfig is the equal but the nodes itself differ, an infoComponentUpdated
// is sent.
// This allows to differ between changes that need to be applied to existing instances
// and infos that relates to scaling (at the moment).
func (c componentNode) Diff(path []string, other node, changes chan<- DiffInfo) {
	otherComponent := other.(componentNode)

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

func diffHierarchy(path []string, newNode, oldNode node, changes chan<- DiffInfo) {
	oldNodes := map[string]node{}
	for _, child := range oldNode.Children() {
		oldNodes[child.Name()] = child
	}

	// Search for nodes that were added or changed
	for _, child := range newNode.Children() {
		name := child.Name()

		oldNode, oldNodeExists := oldNodes[name]
		if oldNodeExists {
			child.Diff(append(path, name), oldNode, changes)

			// This helps us later to determine which nodes were removed
			delete(oldNodes, name)
		} else {
			changes <- DiffInfo{Type: InfoNodeAdded, Name: append(path, name)}
		}
	}

	// Catch all nodes that were removed (the once left do not longer exist in the new config)
	for name, _ := range oldNodes {
		changes <- DiffInfo{Type: InfoNodeRemoved, Name: append(path, name)}
	}
}
