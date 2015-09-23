package userconfig

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type DiffType string

const (
	// DiffInfoServiceNameUpdated
	DiffInfoServiceNameUpdated DiffType = "service-name-updated"

	// DiffInfoComponentAdded
	DiffInfoComponentAdded DiffType = "component-added"

	// DiffInfoComponentRemoved
	DiffInfoComponentRemoved DiffType = "component-removed"

	// DiffInfoComponentUpdated
	DiffInfoComponentUpdated DiffType = "component-updated"

	// NOTE: The following diff info types are currently only for internal usage.
	// We will probably expose them to the user soon, but for now they are just
	// to summarize for a DiffInfoComponentUpdated.

	// DiffInfoComponentImageUpdated
	DiffInfoComponentImageUpdated DiffType = "component-image-updated"

	// DiffInfoComponentEntrypointUpdated
	DiffInfoComponentEntrypointUpdated DiffType = "component-entrypoint-updated"

	// DiffInfoComponentPortsUpdated
	DiffInfoComponentPortsUpdated DiffType = "component-ports-updated"

	// DiffInfoComponentEnvUpdated
	DiffInfoComponentEnvUpdated DiffType = "component-env-updated"

	// DiffInfoComponentVolumesUpdated
	DiffInfoComponentVolumesUpdated DiffType = "component-volumes-updated"

	// DiffInfoComponentArgsUpdated
	DiffInfoComponentArgsUpdated DiffType = "component-args-updated"

	// DiffInfoComponentDomainsUpdated
	DiffInfoComponentDomainsUpdated DiffType = "component-domains-updated"

	// DiffInfoComponentLinksUpdated
	DiffInfoComponentLinksUpdated DiffType = "component-links-updated"

	// DiffInfoComponentExposeUpdated
	DiffInfoComponentExposeUpdated DiffType = "component-expose-updated"

	// scale

	// DiffInfoComponentScalePlacementUpdated
	DiffInfoComponentScalePlacementUpdated DiffType = "component-scale-placement-updated"

	// DiffInfoComponentScaleUp
	DiffInfoComponentScaleUp DiffType = "component-scale-up"

	// DiffInfoComponentScaleDown
	DiffInfoComponentScaleDown DiffType = "component-scale-down"

	// DiffInfoComponentScaleMaxUpdated
	DiffInfoComponentScaleMaxUpdated DiffType = "component-max-updated"

	// DiffInfoComponentPodUpdated
	DiffInfoComponentPodUpdated DiffType = "component-pod-updated"

	//

	// DiffInfoComponentSignalReadyUpdated
	DiffInfoComponentSignalReadyUpdated DiffType = "component-signal-ready-updated"
)

type DiffInfo struct {
	Type DiffType

	Old string
	New string
}

// Action returns a human readable string containing information about what kind of action a
// certain diff type will cause.
func (di DiffInfo) Action() string {
	switch di.Type {
	case DiffInfoServiceNameUpdated:
		return "re-create service"
	case DiffInfoComponentAdded:
		return "add component"
	case DiffInfoComponentUpdated:
		return "update component"
	case DiffInfoComponentRemoved:
		return "remove component"
	case DiffInfoComponentScalePlacementUpdated:
		return "update component"
	case DiffInfoComponentScaleUp:
		return "scale up"
	case DiffInfoComponentScaleDown:
		return "scale down"
	case DiffInfoComponentScaleMaxUpdated:
		return "store component definition"
	default:
		panic(fmt.Sprintf("no action available for unsupported diff type '%s'", di.Type))
	}
}

// Reason returns a human readable string containing information about why a
// certain diff type will cause the related action.
func (di DiffInfo) Reason() string {
	switch di.Type {
	case DiffInfoServiceNameUpdated:
		return "updating service name breaks service discovery"
	case DiffInfoComponentAdded:
		return fmt.Sprintf("component '%s' not found in old definition", di.New)
	case DiffInfoComponentUpdated:
		return fmt.Sprintf("component '%s' changed in new definition", di.New)
	case DiffInfoComponentRemoved:
		return fmt.Sprintf("component '%s' not found in new definition", di.Old)
	case DiffInfoComponentScalePlacementUpdated:
		return fmt.Sprintf("scaling strategy of component '%s' changed in new definition", di.Old)
	case DiffInfoComponentScaleUp:
		return fmt.Sprintf("min scale of component '%s' increased in new definition", di.Old)
	case DiffInfoComponentScaleDown:
		return fmt.Sprintf("min scale of component '%s' decreased in new definition", di.Old)
	case DiffInfoComponentScaleMaxUpdated:
		return fmt.Sprintf("max scale of component '%s' changed in new definition", di.Old)
	default:
		panic(fmt.Sprintf("no reason available for unsupported diff type '%s'", di.Type))
	}
}

// service diff

// ServiceDiff checks the difference between two service definitions. The
// returned list of diff infos can contain the following diff types. Note that
// DiffInfoComponentUpdated is aggregated and details are hidden for the user
// for now.
//   - DiffInfoServiceNameUpdated
//   - DiffInfoComponentAdded
//   - DiffInfoComponentRemoved
//   - DiffInfoComponentUpdated
func ServiceDiff(oldDef, newDef V2AppDefinition) []DiffInfo {
	diffInfos := []DiffInfo{}

	diffInfos = append(diffInfos, diffServiceNameUpdated(oldDef.AppName, newDef.AppName)...)
	diffInfos = append(diffInfos, diffComponentAdded(oldDef.Components, newDef.Components)...)
	diffInfos = append(diffInfos, diffComponentUpdated(oldDef.Components, newDef.Components)...)
	diffInfos = append(diffInfos, diffComponentRemoved(oldDef.Components, newDef.Components)...)

	return diffInfos
}

func diffServiceNameUpdated(oldDef, newDef AppName) []DiffInfo {
	diffInfos := []DiffInfo{}

	if !newDef.Equals(oldDef) {
		diffInfos = append(diffInfos, DiffInfo{
			Type: DiffInfoServiceNameUpdated,
			Old:  oldDef.String(),
			New:  newDef.String(),
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
				Type: DiffInfoComponentAdded,
				New:  newName.String(),
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
				Type: DiffInfoComponentRemoved,
				Old:  oldName.String(),
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
			diffInfos = append(diffInfos, ComponentDiff(*oldComponent, *newComponent, oldName.String())...)
		}
	}

	return diffInfos
}

// component diff

// ComponentDiff checks the difference between two component definitions. The
// returned list of diff infos can contain the following diff types. Note that
// we aggregate all tiff types handled here to create one
// DiffInfoComponentUpdated for the user for now.
//   - DiffInfoComponentImageUpdated
//   - DiffInfoComponentEntrypointUpdated
//   - DiffInfoComponentPortsUpdated
//   - DiffInfoComponentEnvUpdated
//   - DiffInfoComponentVolumesUpdated
//   - DiffInfoComponentArgsUpdated
//   - DiffInfoComponentDomainsUpdated
//   - DiffInfoComponentLinksUpdated
//   - DiffInfoComponentExposeUpdated
//   - DiffInfoComponentScalePlacementUpdated
//   - DiffInfoComponentScaleDown
//   - DiffInfoComponentScaleUp
//   - DiffInfoComponentScaleMaxUpdated
//   - DiffInfoComponentPodUpdated
//   - DiffInfoComponentSignalReadyUpdated
func ComponentDiff(oldDef, newDef ComponentDefinition, componentName string) []DiffInfo {
	publicDiffInfos := []DiffInfo{}  // diff info tracked in detail
	privateDiffInfos := []DiffInfo{} // diff info aggregated to DiffInfoComponentUpdated

	privateDiffInfos = append(privateDiffInfos, diffComponentImage(oldDef, newDef)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentEntrypoint(oldDef, newDef)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentPorts(oldDef, newDef)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentEnv(oldDef, newDef)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentVolumes(oldDef, newDef)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentArgs(oldDef, newDef)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentDomains(oldDef, newDef)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentLinks(oldDef, newDef)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentExpose(oldDef, newDef)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentPod(oldDef, newDef)...)
	privateDiffInfos = append(privateDiffInfos, diffComponentSignalReady(oldDef, newDef)...)

	if len(privateDiffInfos) > 0 {
		publicDiffInfos = append(publicDiffInfos, DiffInfo{
			Type: DiffInfoComponentUpdated,
			Old:  componentName,
			New:  componentName,
		})
	}

	publicDiffInfos = append(publicDiffInfos, diffComponentScale(oldDef, newDef, componentName)...)

	return publicDiffInfos
}

func diffComponentImage(oldDef, newDef ComponentDefinition) []DiffInfo {
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
			Type: DiffInfoComponentImageUpdated,
			Old:  oldImage,
			New:  newImage,
		})
	}

	return diffInfos
}

func diffComponentEntrypoint(oldDef, newDef ComponentDefinition) []DiffInfo {
	diffInfos := []DiffInfo{}

	if oldDef.EntryPoint != newDef.EntryPoint {
		diffInfos = append(diffInfos, DiffInfo{
			Type: DiffInfoComponentEntrypointUpdated,
			Old:  oldDef.EntryPoint,
			New:  newDef.EntryPoint,
		})
	}

	return diffInfos
}

func diffComponentPorts(oldDef, newDef ComponentDefinition) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldPorts := oldDef.Ports.String()
	newPorts := newDef.Ports.String()

	if oldPorts != newPorts {
		diffInfos = append(diffInfos, DiffInfo{
			Type: DiffInfoComponentPortsUpdated,
			Old:  oldPorts,
			New:  newPorts,
		})
	}

	return diffInfos
}

func diffComponentEnv(oldDef, newDef ComponentDefinition) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldEnv := oldDef.Env.String()
	newEnv := newDef.Env.String()

	if oldEnv != newEnv {
		diffInfos = append(diffInfos, DiffInfo{
			Type: DiffInfoComponentEnvUpdated,
			Old:  oldEnv,
			New:  newEnv,
		})
	}

	return diffInfos
}

func diffComponentVolumes(oldDef, newDef ComponentDefinition) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldVolumes := oldDef.Volumes.String()
	newVolumes := newDef.Volumes.String()

	if oldVolumes != newVolumes {
		diffInfos = append(diffInfos, DiffInfo{
			Type: DiffInfoComponentVolumesUpdated,
			Old:  oldVolumes,
			New:  newVolumes,
		})
	}

	return diffInfos
}

func diffComponentArgs(oldDef, newDef ComponentDefinition) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldArgs := strings.Join(oldDef.Args, ", ")
	newArgs := strings.Join(newDef.Args, ", ")

	if oldArgs != newArgs {
		diffInfos = append(diffInfos, DiffInfo{
			Type: DiffInfoComponentArgsUpdated,
			Old:  oldArgs,
			New:  newArgs,
		})
	}

	return diffInfos
}

func diffComponentDomains(oldDef, newDef ComponentDefinition) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldDomains := oldDef.Domains.String()
	newDomains := newDef.Domains.String()

	if oldDomains != newDomains {
		diffInfos = append(diffInfos, DiffInfo{
			Type: DiffInfoComponentLinksUpdated,
			Old:  oldDomains,
			New:  newDomains,
		})
	}

	return diffInfos
}

func diffComponentLinks(oldDef, newDef ComponentDefinition) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldLinks := oldDef.Links.String()
	newLinks := newDef.Links.String()

	if oldLinks != newLinks {
		diffInfos = append(diffInfos, DiffInfo{
			Type: DiffInfoComponentLinksUpdated,
			Old:  oldLinks,
			New:  newLinks,
		})
	}

	return diffInfos
}

func diffComponentExpose(oldDef, newDef ComponentDefinition) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldExpose := oldDef.Expose.String()
	newExpose := newDef.Expose.String()

	if oldExpose != newExpose {
		diffInfos = append(diffInfos, DiffInfo{
			Type: DiffInfoComponentExposeUpdated,
			Old:  oldExpose,
			New:  newExpose,
		})
	}

	return diffInfos
}

// diffComponentScale checks in detail what diff type should be applied between
// oldDef and newDef. The following can be applied.
//   placement changed -> DiffInfoComponentScalePlacementUpdated
//   min decreased     -> DiffInfoComponentScaleDown
//   min increased     -> DiffInfoComponentScaleUp
//   max updated       -> DiffInfoComponentScaleMaxUpdated
func diffComponentScale(oldDef, newDef ComponentDefinition, componentName string) []DiffInfo {
	if oldDef.Scale == nil || newDef.Scale == nil {
		return nil
	}

	diffInfos := []DiffInfo{
		DiffInfo{
			Old: componentName,
			New: componentName,
		},
	}

	// When "placement" changed, we need to update the whole component. This also
	// applies all other definition changes below.
	if !isDefaultPlacement(oldDef.Scale.Placement, newDef.Scale.Placement) && oldDef.Scale.Placement != newDef.Scale.Placement {
		diffInfos[0].Type = DiffInfoComponentScalePlacementUpdated
		return diffInfos
	}

	// When "min" changed, we want to scale a component. This also applies all
	// other definition changes below.
	if oldDef.Scale.Min < newDef.Scale.Min {
		diffInfos[0].Type = DiffInfoComponentScaleUp
		return diffInfos
	}

	// When "min" changed, we want to scale a component. This also applies all
	// other definition changes below.
	if oldDef.Scale.Min > newDef.Scale.Min {
		diffInfos[0].Type = DiffInfoComponentScaleDown
		return diffInfos
	}

	// When "max" changed, we want to update the stored service definition.
	if oldDef.Scale.Max != newDef.Scale.Max {
		diffInfos[0].Type = DiffInfoComponentScaleMaxUpdated
		return diffInfos
	}

	return nil
}

func diffComponentPod(oldDef, newDef ComponentDefinition) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldPod := oldDef.Pod.String()
	newPod := newDef.Pod.String()

	if oldPod != newPod {
		diffInfos = append(diffInfos, DiffInfo{
			Type: DiffInfoComponentPodUpdated,
			Old:  oldPod,
			New:  newPod,
		})
	}

	return diffInfos
}

func diffComponentSignalReady(oldDef, newDef ComponentDefinition) []DiffInfo {
	diffInfos := []DiffInfo{}

	oldPod := strconv.FormatBool(oldDef.SignalReady)
	newPod := strconv.FormatBool(newDef.SignalReady)

	if oldPod != newPod {
		diffInfos = append(diffInfos, DiffInfo{
			Type: DiffInfoComponentSignalReadyUpdated,
			Old:  oldPod,
			New:  newPod,
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
