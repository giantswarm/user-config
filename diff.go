package userconfig

import (
	"sort"
	"strconv"
	"strings"
)

type DiffType string

const (
	// DiffTypeServiceNameUpdated
	DiffTypeServiceNameUpdated DiffType = "service-name-updated"

	// DiffTypeComponentAdded
	DiffTypeComponentAdded DiffType = "component-added"

	// DiffTypeComponentRemoved
	DiffTypeComponentRemoved DiffType = "component-removed"

	// DiffTypeComponentImageUpdated
	DiffTypeComponentImageUpdated DiffType = "component-image-updated"

	// DiffTypeComponentEntrypointUpdated
	DiffTypeComponentEntrypointUpdated DiffType = "component-entrypoint-updated"

	// DiffTypeComponentPortsUpdated
	DiffTypeComponentPortsUpdated DiffType = "component-ports-updated"

	// DiffTypeComponentEnvUpdated
	DiffTypeComponentEnvUpdated DiffType = "component-env-updated"

	// DiffTypeComponentVolumesUpdated
	DiffTypeComponentVolumesUpdated DiffType = "component-volumes-updated"

	// DiffTypeComponentArgsUpdated
	DiffTypeComponentArgsUpdated DiffType = "component-args-updated"

	// DiffTypeComponentDomainsUpdated
	DiffTypeComponentDomainsUpdated DiffType = "component-domains-updated"

	// DiffTypeComponentLinksUpdated
	DiffTypeComponentLinksUpdated DiffType = "component-links-updated"

	// DiffTypeComponentExposeUpdated
	DiffTypeComponentExposeUpdated DiffType = "component-expose-updated"

	// scale

	// DiffTypeComponentScalePlacementUpdated
	DiffTypeComponentScalePlacementUpdated DiffType = "component-scale-placement-updated"

	// DiffTypeComponentScaleMinUpdated
	DiffTypeComponentScaleMinUpdated DiffType = "component-scale-min-updated"

	// DiffTypeComponentScaleMaxUpdated
	DiffTypeComponentScaleMaxUpdated DiffType = "component-scale-max-updated"

	// DiffTypeComponentPodUpdated
	DiffTypeComponentPodUpdated DiffType = "component-pod-updated"

	// DiffTypeComponentSignalReadyUpdated
	DiffTypeComponentSignalReadyUpdated DiffType = "component-signal-ready-updated"
)

type DiffInfo struct {
	Type DiffType

	Component ComponentName

	Key string
	Old string
	New string
}

type DiffInfos []DiffInfo

func (dis DiffInfos) ComponentNames() ComponentNames {
	componentNames := ComponentNames{}

	for _, di := range dis {
		componentNames = append(componentNames, di.Component)
	}

	return componentNames
}

// service diff

// ServiceDiff checks the difference between two service definitions. The
// returned list of diff infos can contain the following diff types.
//   - DiffTypeServiceNameUpdated
//   - DiffTypeComponentAdded
//   - DiffTypeComponentRemoved
func ServiceDiff(oldDef, newDef V2AppDefinition) DiffInfos {
	diffInfos := DiffInfos{}

	diffInfos = append(diffInfos, diffServiceNameUpdated(oldDef.AppName, newDef.AppName)...)
	diffInfos = append(diffInfos, diffComponentAdded(oldDef.Components, newDef.Components)...)
	diffInfos = append(diffInfos, diffComponentUpdated(oldDef.Components, newDef.Components)...)
	diffInfos = append(diffInfos, diffComponentRemoved(oldDef.Components, newDef.Components)...)

	return diffInfos
}

func DiffInfosByType(diffInfos DiffInfos, t DiffType) DiffInfos {
	newDiffInfos := DiffInfos{}

	for _, diffInfo := range diffInfos {
		if diffInfo.Type == t {
			newDiffInfos = append(newDiffInfos, diffInfo)
		}
	}

	return newDiffInfos
}

func FilterDiffType(diffInfos DiffInfos, diffType DiffType) DiffInfos {
	newDiffInfos := DiffInfos{}

	for _, di := range diffInfos {
		if di.Type == diffType {
			continue
		}

		newDiffInfos = append(newDiffInfos, di)
	}

	return newDiffInfos
}

func diffServiceNameUpdated(oldName, newName AppName) DiffInfos {
	diffInfos := DiffInfos{}

	if !newName.Equals(oldName) {
		diffInfos = append(diffInfos, DiffInfo{
			Type: DiffTypeServiceNameUpdated,
			Key:  "name",
			Old:  oldName.String(),
			New:  newName.String(),
		})
	}

	return diffInfos
}

func diffComponentAdded(oldDef, newDef ComponentDefinitions) DiffInfos {
	diffInfos := DiffInfos{}

	for _, orderedName := range orderedComponentKeys(newDef) {
		newName := ComponentName(orderedName)

		if _, ok := oldDef[newName]; !ok {
			diffInfos = append(diffInfos, DiffInfo{
				Type:      DiffTypeComponentAdded,
				Component: newName,
				New:       newName.String(),
			})
		}
	}

	return diffInfos
}

func diffComponentRemoved(oldDef, newDef ComponentDefinitions) DiffInfos {
	diffInfos := DiffInfos{}

	for _, orderedName := range orderedComponentKeys(oldDef) {
		oldName := ComponentName(orderedName)

		if _, ok := newDef[oldName]; !ok {
			diffInfos = append(diffInfos, DiffInfo{
				Type:      DiffTypeComponentRemoved,
				Component: oldName,
				Old:       oldName.String(),
			})
		}
	}

	return diffInfos
}

func diffComponentUpdated(oldDef, newDef ComponentDefinitions) DiffInfos {
	diffInfos := DiffInfos{}

	for _, orderedName := range orderedComponentKeys(oldDef) {
		oldName := ComponentName(orderedName)
		oldComponent := oldDef[oldName]

		if newComponent, ok := newDef[oldName]; ok {
			diffInfos = append(diffInfos, ComponentDiff(*oldComponent, *newComponent, oldName)...)
		}
	}

	return diffInfos
}

// component diff

// ComponentDiff checks the difference between two component definitions. The
// returned list of diff infos can contain the following diff types.
//   - DiffTypeComponentImageUpdated
//   - DiffTypeComponentEntrypointUpdated
//   - DiffTypeComponentPortsUpdated
//   - DiffTypeComponentEnvUpdated
//   - DiffTypeComponentVolumesUpdated
//   - DiffTypeComponentArgsUpdated
//   - DiffTypeComponentDomainsUpdated
//   - DiffTypeComponentLinksUpdated
//   - DiffTypeComponentExposeUpdated
//   - DiffTypeComponentScalePlacementUpdated
//   - DiffTypeComponentScaleMinUpdated
//   - DiffTypeComponentScaleMaxUpdated
//   - DiffTypeComponentPodUpdated
//   - DiffTypeComponentSignalReadyUpdated
func ComponentDiff(oldDef, newDef ComponentDefinition, componentName ComponentName) DiffInfos {
	diffInfos := DiffInfos{} // diff info tracked in detail

	diffInfos = append(diffInfos, diffComponentImage(oldDef, newDef, componentName)...)
	diffInfos = append(diffInfos, diffComponentEntrypoint(oldDef, newDef, componentName)...)
	diffInfos = append(diffInfos, diffComponentPorts(oldDef, newDef, componentName)...)
	diffInfos = append(diffInfos, diffComponentEnv(oldDef, newDef, componentName)...)
	diffInfos = append(diffInfos, diffComponentVolumes(oldDef, newDef, componentName)...)
	diffInfos = append(diffInfos, diffComponentArgs(oldDef, newDef, componentName)...)
	diffInfos = append(diffInfos, diffComponentDomains(oldDef, newDef, componentName)...)
	diffInfos = append(diffInfos, diffComponentLinks(oldDef, newDef, componentName)...)
	diffInfos = append(diffInfos, diffComponentExpose(oldDef, newDef, componentName)...)
	diffInfos = append(diffInfos, diffComponentPod(oldDef, newDef, componentName)...)
	diffInfos = append(diffInfos, diffComponentSignalReady(oldDef, newDef, componentName)...)
	diffInfos = append(diffInfos, diffComponentScale(oldDef, newDef, componentName)...)

	return diffInfos
}

func diffComponentImage(oldDef, newDef ComponentDefinition, componentName ComponentName) DiffInfos {
	diffInfos := DiffInfos{}

	// NOTE: Components don't need to have an Image.
	var oldImage, newImage string
	if oldDef.Image != nil {
		oldImage = oldDef.Image.String()
	}
	if newDef.Image != nil {
		newImage = newDef.Image.String()
	}

	if oldImage != newImage {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentImageUpdated,
			Key:       "image",
			Component: componentName,
			Old:       oldImage,
			New:       newImage,
		})
	}

	return diffInfos
}

func diffComponentEntrypoint(oldDef, newDef ComponentDefinition, componentName ComponentName) DiffInfos {
	diffInfos := DiffInfos{}

	if oldDef.EntryPoint != newDef.EntryPoint {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentEntrypointUpdated,
			Key:       "entrypoint",
			Component: componentName,
			Old:       oldDef.EntryPoint,
			New:       newDef.EntryPoint,
		})
	}

	return diffInfos
}

func diffComponentPorts(oldDef, newDef ComponentDefinition, componentName ComponentName) DiffInfos {
	diffInfos := DiffInfos{}

	// TODO this needs to be more fine grained
	oldPorts := oldDef.Ports.String()
	newPorts := newDef.Ports.String()

	if oldPorts != newPorts {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentPortsUpdated,
			Key:       "ports",
			Component: componentName,
			Old:       oldPorts,
			New:       newPorts,
		})
	}

	return diffInfos
}

func diffComponentEnv(oldDef, newDef ComponentDefinition, componentName ComponentName) DiffInfos {
	diffInfos := DiffInfos{}

	// TODO this needs to be more fine grained
	oldEnv := oldDef.Env.String()
	newEnv := newDef.Env.String()

	if oldEnv != newEnv {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentEnvUpdated,
			Key:       "env",
			Component: componentName,
			Old:       oldEnv,
			New:       newEnv,
		})
	}

	return diffInfos
}

func diffComponentVolumes(oldDef, newDef ComponentDefinition, componentName ComponentName) DiffInfos {
	diffInfos := DiffInfos{}

	// TODO this needs to be more fine grained
	oldVolumes := oldDef.Volumes.String()
	newVolumes := newDef.Volumes.String()

	if oldVolumes != newVolumes {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentVolumesUpdated,
			Key:       "volumes",
			Component: componentName,
			Old:       oldVolumes,
			New:       newVolumes,
		})
	}

	return diffInfos
}

func diffComponentArgs(oldDef, newDef ComponentDefinition, componentName ComponentName) DiffInfos {
	diffInfos := DiffInfos{}

	// TODO this needs to be more fine grained
	oldArgs := strings.Join(oldDef.Args, ", ")
	newArgs := strings.Join(newDef.Args, ", ")

	if oldArgs != newArgs {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentArgsUpdated,
			Key:       "args",
			Component: componentName,
			Old:       oldArgs,
			New:       newArgs,
		})
	}

	return diffInfos
}

func diffComponentDomains(oldDef, newDef ComponentDefinition, componentName ComponentName) DiffInfos {
	diffInfos := DiffInfos{}

	// TODO this needs to be more fine grained
	oldDomains := oldDef.Domains.String()
	newDomains := newDef.Domains.String()

	if oldDomains != newDomains {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentLinksUpdated,
			Key:       "domains",
			Component: componentName,
			Old:       oldDomains,
			New:       newDomains,
		})
	}

	return diffInfos
}

func diffComponentLinks(oldDef, newDef ComponentDefinition, componentName ComponentName) DiffInfos {
	diffInfos := DiffInfos{}

	// TODO this needs to be more fine grained
	oldLinks := oldDef.Links.String()
	newLinks := newDef.Links.String()

	if oldLinks != newLinks {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentLinksUpdated,
			Key:       "links",
			Component: componentName,
			Old:       oldLinks,
			New:       newLinks,
		})
	}

	return diffInfos
}

func diffComponentExpose(oldDef, newDef ComponentDefinition, componentName ComponentName) DiffInfos {
	diffInfos := DiffInfos{}

	oldExpose := oldDef.Expose.String()
	newExpose := newDef.Expose.String()

	if oldExpose != newExpose {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentExposeUpdated,
			Key:       "expose",
			Component: componentName,
			Old:       oldExpose,
			New:       newExpose,
		})
	}

	return diffInfos
}

// diffComponentScale checks in detail what diff type should be applied between
// oldDef and newDef. The following can be applied.
//   - DiffTypeComponentScalePlacementUpdated
//   - DiffTypeComponentScaleMinUpdated
//   - DiffTypeComponentScaleMaxUpdated
func diffComponentScale(oldDef, newDef ComponentDefinition, componentName ComponentName) DiffInfos {
	oldScaleDef := &ScaleDefinition{}
	if oldDef.Scale != nil {
		oldScaleDef = oldDef.Scale
	}

	newScaleDef := &ScaleDefinition{}
	if newDef.Scale != nil {
		newScaleDef = newDef.Scale
	}

	diffInfos := DiffInfos{}

	if !isDefaultPlacement(oldScaleDef.Placement, newScaleDef.Placement) && oldScaleDef.Placement != newScaleDef.Placement {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentScalePlacementUpdated,
			Key:       "scale.placement",
			Component: componentName,
			Old:       string(oldScaleDef.Placement),
			New:       string(newScaleDef.Placement),
		})
	}

	if oldScaleDef.Min != newScaleDef.Min {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentScaleMinUpdated,
			Key:       "scale.min",
			Component: componentName,
			Old:       strconv.Itoa(oldScaleDef.Min),
			New:       strconv.Itoa(newScaleDef.Min),
		})
	}

	if oldScaleDef.Max != newScaleDef.Max {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentScaleMaxUpdated,
			Key:       "scale.max",
			Component: componentName,
			Old:       strconv.Itoa(oldScaleDef.Max),
			New:       strconv.Itoa(newScaleDef.Max),
		})
	}

	return diffInfos
}

func diffComponentPod(oldDef, newDef ComponentDefinition, componentName ComponentName) DiffInfos {
	diffInfos := DiffInfos{}

	oldPod := oldDef.Pod.String()
	newPod := newDef.Pod.String()

	if oldPod != newPod {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentPodUpdated,
			Key:       "pod",
			Component: componentName,
			Old:       oldPod,
			New:       newPod,
		})
	}

	return diffInfos
}

func diffComponentSignalReady(oldDef, newDef ComponentDefinition, componentName ComponentName) DiffInfos {
	diffInfos := DiffInfos{}

	oldPod := strconv.FormatBool(oldDef.SignalReady)
	newPod := strconv.FormatBool(newDef.SignalReady)

	if oldPod != newPod {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentSignalReadyUpdated,
			Key:       "signal-ready",
			Component: componentName,
			Old:       oldPod,
			New:       newPod,
		})
	}

	return diffInfos
}

// helper

// orderedComponentKeys creates a ordered list of component names, based on the
// provided component map.
func orderedComponentKeys(defs ComponentDefinitions) []string {
	keys := []string{}

	for name, _ := range defs {
		keys = append(keys, name.String())
	}
	sort.Strings(keys)

	return keys
}

func isDefaultPlacement(oldPlacement, newPlacement Placement) bool {
	return ((oldPlacement == "" || oldPlacement == DefaultPlacement) && (newPlacement == "" || newPlacement == DefaultPlacement))
}
