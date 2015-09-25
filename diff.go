package userconfig

import (
	"fmt"
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

	// DiffTypeComponentUpdated
	DiffTypeComponentUpdated DiffType = "component-updated"

	// NOTE: The following diff info types are currently only for internal usage.
	// We will probably expose them to the user soon, but for now they are just
	// to summarize for a DiffTypeComponentUpdated.

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

	// DiffTypeComponentScaleMinDecreased
	DiffTypeComponentScaleMinDecreased DiffType = "component-scale-min-decreased"

	// DiffTypeComponentScaleMinIncreased
	DiffTypeComponentScaleMinIncreased DiffType = "component-scale-min-increased"

	// DiffTypeComponentScaleMaxDecreased
	DiffTypeComponentScaleMaxDecreased DiffType = "component-scale-max-decreased"

	// DiffTypeComponentScaleMaxIncreased
	DiffTypeComponentScaleMaxIncreased DiffType = "component-scale-max-increased"

	//

	// DiffTypeComponentPodUpdated
	DiffTypeComponentPodUpdated DiffType = "component-pod-updated"

	// DiffTypeComponentSignalReadyUpdated
	DiffTypeComponentSignalReadyUpdated DiffType = "component-signal-ready-updated"
)

type DiffInfo struct {
	Type DiffType

	Component ComponentName

	// TODO in case we don't use the DiffInfo structure for displaying a plan to
	// the user, we need to move Action and Reason out of this.
	Action string
	Reason string

	Old string
	New string
}

// service diff

// ServiceDiff checks the difference between two service definitions. The
// returned list of diff infos can contain the following diff types. Note that
// DiffTypeComponentUpdated is aggregated and details are hidden for the user
// for now.
//   - DiffTypeServiceNameUpdated
//   - DiffTypeComponentAdded
//   - DiffTypeComponentRemoved
//   - DiffTypeComponentUpdated
func ServiceDiff(oldDef, newDef V2AppDefinition) []DiffInfo {
	diffInfos := []DiffInfo{}

	diffInfos = append(diffInfos, diffServiceNameUpdated(oldDef.AppName, newDef.AppName)...)
	diffInfos = append(diffInfos, diffComponentAdded(oldDef.Components, newDef.Components)...)
	diffInfos = append(diffInfos, diffComponentUpdated(oldDef.Components, newDef.Components)...)
	diffInfos = append(diffInfos, diffComponentRemoved(oldDef.Components, newDef.Components)...)

	return diffInfos
}

func diffServiceNameUpdated(oldName, newName AppName) []DiffInfo {
	diffInfos := []DiffInfo{}

	if !newName.Equals(oldName) {
		diffInfos = append(diffInfos, DiffInfo{
			Type:   DiffTypeServiceNameUpdated,
			Action: "re-create service",
			Reason: "updating service name breaks service discovery",
			Old:    oldName.String(),
			New:    newName.String(),
		})
	}

	return diffInfos
}

func diffComponentAdded(oldDef, newDef ComponentDefinitions) []DiffInfo {
	diffInfos := []DiffInfo{}

	for _, orderedName := range orderedComponentKeys(newDef) {
		newName := ComponentName(orderedName)

		if _, ok := oldDef[newName]; !ok {
			diffInfos = append(diffInfos, DiffInfo{
				Type:      DiffTypeComponentAdded,
				Component: newName,
				Action:    "add component",
				Reason:    fmt.Sprintf("component '%s' not found in old definition", newName),
				New:       newName.String(),
			})
		}
	}

	return diffInfos
}

func diffComponentRemoved(oldDef, newDef ComponentDefinitions) []DiffInfo {
	diffInfos := []DiffInfo{}

	for _, orderedName := range orderedComponentKeys(oldDef) {
		oldName := ComponentName(orderedName)

		if _, ok := newDef[oldName]; !ok {
			diffInfos = append(diffInfos, DiffInfo{
				Type:      DiffTypeComponentRemoved,
				Component: oldName,
				Action:    "remove component",
				Reason:    fmt.Sprintf("component '%s' not found in new definition", oldName),
				Old:       oldName.String(),
			})
		}
	}

	return diffInfos
}

func diffComponentUpdated(oldDef, newDef ComponentDefinitions) []DiffInfo {
	diffInfos := []DiffInfo{}

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
// returned list of diff infos can contain the following diff types. Note that
// we aggregate all tiff types handled here to create one
// DiffTypeComponentUpdated for the user for now.
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
//   - DiffTypeComponentScaleDown
//   - DiffTypeComponentScaleUp
//   - DiffTypeComponentScaleMaxUpdated
//   - DiffTypeComponentPodUpdated
//   - DiffTypeComponentSignalReadyUpdated
func ComponentDiff(oldDef, newDef ComponentDefinition, componentName ComponentName) []DiffInfo {
	publicDiffInfos := []DiffInfo{}  // diff info tracked in detail
	privateDiffInfos := []DiffInfo{} // diff info aggregated to DiffTypeComponentUpdated

	privateDiffInfos = append(privateDiffInfos, diffComponentImage(oldDef, newDef, componentName)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentEntrypoint(oldDef, newDef, componentName)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentPorts(oldDef, newDef, componentName)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentEnv(oldDef, newDef, componentName)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentVolumes(oldDef, newDef, componentName)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentArgs(oldDef, newDef, componentName)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentDomains(oldDef, newDef, componentName)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentLinks(oldDef, newDef, componentName)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentExpose(oldDef, newDef, componentName)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentPod(oldDef, newDef, componentName)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentSignalReady(oldDef, newDef, componentName)...)

	if len(privateDiffInfos) > 0 {
		publicDiffInfos = append(publicDiffInfos, DiffInfo{
			Type:      DiffTypeComponentUpdated,
			Component: componentName,
			Action:    "update component",
			Reason:    fmt.Sprintf("component '%s' changed in new definition", componentName),
		})
	}

	publicDiffInfos = append(publicDiffInfos, diffComponentScale(oldDef, newDef, componentName)...)

	return publicDiffInfos
}

func diffComponentImage(oldDef, newDef ComponentDefinition, componentName ComponentName) []DiffInfo {
	diffInfos := []DiffInfo{}

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
			Component: componentName,
			Old:       oldImage,
			New:       newImage,
		})
	}

	return diffInfos
}

func diffComponentEntrypoint(oldDef, newDef ComponentDefinition, componentName ComponentName) []DiffInfo {
	diffInfos := []DiffInfo{}

	if oldDef.EntryPoint != newDef.EntryPoint {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentEntrypointUpdated,
			Component: componentName,
			Old:       oldDef.EntryPoint,
			New:       newDef.EntryPoint,
		})
	}

	return diffInfos
}

func diffComponentPorts(oldDef, newDef ComponentDefinition, componentName ComponentName) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldPorts := oldDef.Ports.String()
	newPorts := newDef.Ports.String()

	if oldPorts != newPorts {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentPortsUpdated,
			Component: componentName,
			Old:       oldPorts,
			New:       newPorts,
		})
	}

	return diffInfos
}

func diffComponentEnv(oldDef, newDef ComponentDefinition, componentName ComponentName) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldEnv := oldDef.Env.String()
	newEnv := newDef.Env.String()

	if oldEnv != newEnv {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentEnvUpdated,
			Component: componentName,
			Old:       oldEnv,
			New:       newEnv,
		})
	}

	return diffInfos
}

func diffComponentVolumes(oldDef, newDef ComponentDefinition, componentName ComponentName) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldVolumes := oldDef.Volumes.String()
	newVolumes := newDef.Volumes.String()

	if oldVolumes != newVolumes {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentVolumesUpdated,
			Component: componentName,
			Old:       oldVolumes,
			New:       newVolumes,
		})
	}

	return diffInfos
}

func diffComponentArgs(oldDef, newDef ComponentDefinition, componentName ComponentName) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldArgs := strings.Join(oldDef.Args, ", ")
	newArgs := strings.Join(newDef.Args, ", ")

	if oldArgs != newArgs {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentArgsUpdated,
			Component: componentName,
			Old:       oldArgs,
			New:       newArgs,
		})
	}

	return diffInfos
}

func diffComponentDomains(oldDef, newDef ComponentDefinition, componentName ComponentName) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldDomains := oldDef.Domains.String()
	newDomains := newDef.Domains.String()

	if oldDomains != newDomains {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentLinksUpdated,
			Component: componentName,
			Old:       oldDomains,
			New:       newDomains,
		})
	}

	return diffInfos
}

func diffComponentLinks(oldDef, newDef ComponentDefinition, componentName ComponentName) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldLinks := oldDef.Links.String()
	newLinks := newDef.Links.String()

	if oldLinks != newLinks {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentLinksUpdated,
			Component: componentName,
			Old:       oldLinks,
			New:       newLinks,
		})
	}

	return diffInfos
}

func diffComponentExpose(oldDef, newDef ComponentDefinition, componentName ComponentName) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldExpose := oldDef.Expose.String()
	newExpose := newDef.Expose.String()

	if oldExpose != newExpose {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentExposeUpdated,
			Component: componentName,
			Old:       oldExpose,
			New:       newExpose,
		})
	}

	return diffInfos
}

// diffComponentScale checks in detail what diff type should be applied between
// oldDef and newDef. The following can be applied.
//   placement changed -> DiffTypeComponentScalePlacementUpdated
//   min decreased     -> DiffTypeComponentScaleMinDecreased
//   min increased     -> DiffTypeComponentScaleMinIncreased
//   max decreased     -> DiffTypeComponentScaleMaxDecreased
//   max increased     -> DiffTypeComponentScaleMaxIncreased
func diffComponentScale(oldDef, newDef ComponentDefinition, componentName ComponentName) []DiffInfo {
	if oldDef.Scale == nil || newDef.Scale == nil {
		return nil
	}

	diffInfos := []DiffInfo{}

	if !isDefaultPlacement(oldDef.Scale.Placement, newDef.Scale.Placement) && oldDef.Scale.Placement != newDef.Scale.Placement {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentScalePlacementUpdated,
			Component: componentName,
			Action:    "update component",
			Reason:    fmt.Sprintf("scaling strategy of component '%s' changed in new definition", componentName),
			Old:       oldDef.Scale.String(),
			New:       newDef.Scale.String(),
		})
	}

	if oldDef.Scale.Min > newDef.Scale.Min {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentScaleMinDecreased,
			Component: componentName,
			Action:    "store component definition",
			Reason:    fmt.Sprintf("min scale of component '%s' decreased in new definition", componentName),
			Old:       oldDef.Scale.String(),
			New:       newDef.Scale.String(),
		})
	}

	if oldDef.Scale.Min < newDef.Scale.Min {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentScaleMinIncreased,
			Component: componentName,
			Action:    "", // we need to decide server side what action to apply
			Reason:    "", // we need to decide server side what action to apply
			Old:       oldDef.Scale.String(),
			New:       newDef.Scale.String(),
		})
	}

	if oldDef.Scale.Max > newDef.Scale.Max {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentScaleMaxDecreased,
			Component: componentName,
			Action:    "", // we need to decide server side what action to apply
			Reason:    "", // we need to decide server side what action to apply
			Old:       oldDef.Scale.String(),
			New:       newDef.Scale.String(),
		})
	}

	if oldDef.Scale.Max < newDef.Scale.Max {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentScaleMaxIncreased,
			Component: componentName,
			Action:    "store component definition",
			Reason:    fmt.Sprintf("max scale of component '%s' increased in new definition", componentName),
			Old:       oldDef.Scale.String(),
			New:       newDef.Scale.String(),
		})
	}

	return diffInfos
}

func diffComponentPod(oldDef, newDef ComponentDefinition, componentName ComponentName) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldPod := oldDef.Pod.String()
	newPod := newDef.Pod.String()

	if oldPod != newPod {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentPodUpdated,
			Component: componentName,
			Old:       oldPod,
			New:       newPod,
		})
	}

	return diffInfos
}

func diffComponentSignalReady(oldDef, newDef ComponentDefinition, componentName ComponentName) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldPod := strconv.FormatBool(oldDef.SignalReady)
	newPod := strconv.FormatBool(newDef.SignalReady)

	if oldPod != newPod {
		diffInfos = append(diffInfos, DiffInfo{
			Type:      DiffTypeComponentSignalReadyUpdated,
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

func filterDiffType(diffInfos []DiffInfo, diffType DiffType) []DiffInfo {
	newDiffInfos := []DiffInfo{}

	for _, di := range diffInfos {
		if di.Type == diffType {
			continue
		}

		newDiffInfos = append(newDiffInfos, di)
	}

	return newDiffInfos
}

func isDefaultPlacement(oldPlacement, newPlacement Placement) bool {
	return ((oldPlacement == "" || oldPlacement == DefaultPlacement) && (newPlacement == "" || newPlacement == DefaultPlacement))
}
