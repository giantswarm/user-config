package userconfig

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/giantswarm/generic-types-go"
	"github.com/juju/errgo"
	"github.com/kr/pretty"
)

type appConfigCopy AppDefinition

func CheckForUnknownFields(b []byte, ac *AppDefinition) error {
	var clean appConfigCopy
	if err := json.Unmarshal(b, &clean); err != nil {
		return Mask(err)
	}

	cleanBytes, err := json.Marshal(clean)
	if err != nil {
		return Mask(err)
	}

	var dirtyMap map[string]interface{}
	if err := json.Unmarshal(b, &dirtyMap); err != nil {
		return Mask(err)
	}
	// Normalize fields to common format
	normalizeEnv(dirtyMap)
	normalizeVolumeSizes(dirtyMap)

	var cleanMap map[string]interface{}
	if err := json.Unmarshal(cleanBytes, &cleanMap); err != nil {
		return Mask(err)
	}

	diff := pretty.Diff(dirtyMap, cleanMap)
	for _, v := range diff {
		*ac = AppDefinition{}

		field := strings.Split(v, ":")
		return errgo.WithCausef(nil, UnknownJSONFieldError, "Cannot parse app config. Unknown field '%s' detected.", field[0])
	}

	return nil
}

// normalizeEnv normalizes all struct "env" elements under service/component to its natural array format.
// This normalization function is expected to normalize "valid" data and passthrough everything else.
func normalizeEnv(config map[string]interface{}) {
	services := getArrayEntry(config, "services")
	if services == nil {
		// No services element
		return
	}
	for _, service := range services {
		serviceMap, ok := service.(map[string]interface{})
		if !ok {
			continue
		}
		components := getArrayEntry(serviceMap, "components")
		if components == nil {
			// No components element
			continue
		}
		for _, component := range components {
			componentMap, ok := component.(map[string]interface{})
			if !ok {
				continue
			}
			env, ok := componentMap["env"]
			if !ok {
				continue
			}
			envMap, ok := env.(map[string]interface{})
			if !ok {
				// Not of the map type
				continue
			}
			// 'env' is of map type, normalize it to an array
			// Sort the keys first so the outcome it always the same
			keys := []string{}
			for k, _ := range envMap {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			list := []interface{}{}
			for _, k := range keys {
				v := envMap[k]
				list = append(list, fmt.Sprintf("%s=%s", k, v))
			}
			componentMap["env"] = list
		}
	}
}

// normalizeVolumeSizes normalizes all volume sizes to it's normalized format of "number GB"
// This normalization function is expected to normalize "valid" data and passthrough everything else.
func normalizeVolumeSizes(config map[string]interface{}) {
	services := getArrayEntry(config, "services")
	if services == nil {
		// No services element
		return
	}
	for _, service := range services {
		serviceMap, ok := service.(map[string]interface{})
		if !ok {
			continue
		}
		components := getArrayEntry(serviceMap, "components")
		if components == nil {
			// No components element
			continue
		}
		for _, component := range components {
			componentMap, ok := component.(map[string]interface{})
			if !ok {
				continue
			}
			volumes := getArrayEntry(componentMap, "volumes")
			if volumes == nil {
				// No volumes element
				continue
			}
			for _, volume := range volumes {
				volumeMap, ok := volume.(map[string]interface{})
				if !ok {
					continue
				}
				sizeRaw, ok := volumeMap["size"]
				if !ok {
					// Size not found
					continue
				}
				size, ok := sizeRaw.(string)
				if !ok {
					// size is not a string
					continue
				}
				// Parse volume size
				var volumeSize VolumeSize
				// Marshal size string to json
				data, err := json.Marshal(size)
				if err != nil {
					continue
				}
				// Try to unmarshal volume size
				if err := volumeSize.UnmarshalJSON(data); err != nil {
					// Not valid format
					continue
				}
				// Use normalized format
				volumeMap["size"] = string(volumeSize)
			}
		}
	}
}

// getArrayEntry tries to get an entry in the given map that is an array of objects.
func getArrayEntry(config map[string]interface{}, key string) []interface{} {
	entry, ok := config[key]
	if !ok {
		// No key element found
		return nil
	}

	entryArr, ok := entry.([]interface{})
	if !ok {
		// entry not right type
		return nil
	}

	return entryArr
}

// validate performs semantic validations of this AppDefinition.
// Return the first possible error.
func (ad *AppDefinition) validate() error {
	for _, s := range ad.Services {
		if err := s.validate(); err != nil {
			return Mask(err)
		}
	}
	if err := ad.validatePods(); err != nil {
		return Mask(err)
	}

	return nil
}

// validate performs semantic validations of this ServiceConfig.
// Return the first possible error.
func (sc *ServiceConfig) validate() error {
	for _, c := range sc.Components {
		if err := c.validate(); err != nil {
			return Mask(err)
		}
		// Check volume refs
		for _, v := range c.Volumes {
			if err := v.validateVolumeRefs(sc, &c); err != nil {
				return Mask(err)
			}
		}
		// Check for duplicate mount points
		if err := c.validateUniqueMountPoints(sc); err != nil {
			return Mask(err)
		}
	}

	// Check for duplicate exposed ports in pods
	if err := sc.validateUniquePortsInPods(); err != nil {
		return Mask(err)
	}

	// Check dependencies in pods
	if err := sc.validateUniqueDependenciesInPods(); err != nil {
		return Mask(err)
	}

	// Check scaling policies in pods
	if err := sc.validateScalingPolicyInPods(); err != nil {
		return Mask(err)
	}

	return nil
}

// validate performs semantic validations of this ComponentConfig.
// Return the first possible error.
func (cc *ComponentConfig) validate() error {
	// Check volumes
	for _, v := range cc.Volumes {
		if err := v.validate(); err != nil {
			return err
		}
	}

	for d, _ := range cc.Domains {
		if err := d.Validate(); err != nil {
			return Mask(err)
		}
	}

	// No errors found
	return nil
}

type podInfoCounter struct {
	ServiceName string
	Count       int
}

// validatePods checks that
// - pods do not cross service boundaries.
// - pods must be used in more than 1 component.
func (ad *AppDefinition) validatePods() error {
	pod2info := make(map[string]*podInfoCounter)
	for _, s := range ad.Services {
		for _, c := range s.Components {
			pn := c.PodName
			if pn != "" {
				info, ok := pod2info[pn]
				if !ok {
					// First occurrence of the pod
					pod2info[pn] = &podInfoCounter{s.ServiceName, 1}
				} else {
					// Found earlier use of pod name
					if info.ServiceName != s.ServiceName {
						// Pod is used in different services
						return errgo.WithCausef(nil, CrossServicePodError, "Cannot parse app config. Pod '%s' is used in multiple services.", pn)
					}
					// Increase counter
					info.Count++
				}
			}
		}
	}
	// Test counters
	for pn, info := range pod2info {
		if info.Count == 1 {
			// Pod is used only once
			return errgo.WithCausef(nil, PodUsedOnlyOnceError, "Cannot parse app config. Pod '%s' is used in only 1 component.", pn)
		}
	}
	return nil
}

// validate validates the settings of this VolumeConfig.
// Valid combinations:
// - Option1: Path & Size set, everything else empty
// - Option 2: VolumesFrom set, everything else empty
// - Option 3: VolumeFrom, VolumePath set, Path optionally set, everything else empty
func (vc *VolumeConfig) validate() error {
	// Option 1
	if vc.Path != "" && !vc.Size.Empty() {
		if vc.VolumesFrom != "" {
			return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse volume config. Volumes-from for path '%s' should be empty.", vc.Path)
		}
		if vc.VolumeFrom != "" {
			return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse volume config. Volume-from for path '%s' should be empty.", vc.Path)
		}
		if vc.VolumePath != "" {
			return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse volume config. Volume-path for path '%s' should be empty.", vc.Path)
		}
		return nil
	}
	// Option 2
	if vc.VolumesFrom != "" {
		if vc.Path != "" {
			return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse volume config. Path for volumes-from '%s' should be empty.", vc.VolumesFrom)
		}
		if !vc.Size.Empty() {
			return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse volume config. Size for volumes-from '%s' should be empty.", vc.VolumesFrom)
		}
		if vc.VolumeFrom != "" {
			return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse volume config. Volume-from for volumes-from '%s' should be empty.", vc.VolumesFrom)
		}
		if vc.VolumePath != "" {
			return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse volume config. Volume-path for volumes-from '%s' should be empty.", vc.VolumesFrom)
		}
		return nil
	}
	// Option 3
	if vc.VolumeFrom != "" && vc.VolumePath != "" {
		// Path is optional

		if !vc.Size.Empty() {
			return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse volume config. Size for volume-from '%s' should be empty.", vc.VolumeFrom)
		}
		if vc.VolumesFrom != "" {
			return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse volume config. Volumes-from for volume-from '%s' should be empty.", vc.VolumeFrom)
		}
		if vc.VolumePath == "" {
			return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse volume config. Volume-path for volume-from '%s' should not be empty.", vc.VolumeFrom)
		}
		return nil
	}

	// No valid option detected.
	return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse volume config. Path, volume-path or volumes-path must be set. %#v", vc)
}

// validateVolumeRefs checks the existance of reference names in the given volume config.
func (vc *VolumeConfig) validateVolumeRefs(service *ServiceConfig, containingComponent *ComponentConfig) error {
	compName := vc.VolumesFrom
	if compName == "" {
		compName = vc.VolumeFrom
	}
	if compName == "" {
		// No references, all ok
		return nil
	}

	// Check that the component name (volume-from or volumes-from) is not the containing component
	if compName == containingComponent.ComponentName {
		return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse volume config. Cannot refer to own component '%s'.", compName)
	}
	// Another component is referenced, we should be in a pod
	pn := containingComponent.PodName
	if pn == "" {
		return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse volume config. Cannot refer to another component '%s' without a pod declaration.", compName)
	}
	// Find the other component name
	other := service.findComponent(compName)
	if other != nil {
		// Found other component
		// Check matching pod
		if pn != other.PodName {
			return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse volume config. Cannot refer to another component '%s' without a matching pod declaration.", compName)
		}
		// Check matching "volume-path"
		if vc.VolumePath != "" {
			found := false
			for _, v := range other.Volumes {
				if v.Path == vc.VolumePath {
					// Found it
					found = true
				}
			}
			if !found {
				return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse volume config. Cannot find path '%s' on component '%s'.", vc.VolumePath, compName)
			}
		}
		// all ok
		return nil
	}

	// Not found
	return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse volume config. Cannot find referenced component '%s'.", compName)
}

// validateUniqueMountPoints checks that there are no duplicate volume mounts
func (cc *ComponentConfig) validateUniqueMountPoints(service *ServiceConfig) error {
	mountPoints := make(map[string]string)
	for _, v := range cc.Volumes {
		var paths []string
		if v.Path != "" {
			paths = []string{v.Path}
		} else if v.VolumeFrom != "" {
			paths = []string{v.VolumePath}
		} else if v.VolumesFrom != "" {
			other := service.findComponent(v.VolumesFrom)
			if other == nil {
				return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse app config. Cannot find referenced component '%s'.", v.VolumesFrom)
			}
			visitedComponents := make(map[string]string)
			var err error
			paths, err = other.getAllMountPoints(service, visitedComponents)
			if err != nil {
				return err
			}
		} else {
			return errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse app config. Missing path in component '%s'.", cc.ComponentName)
		}
		for _, p := range paths {
			if _, ok := mountPoints[p]; ok {
				// Found duplicate mount point
				return errgo.WithCausef(nil, DuplicateVolumePathError, "Cannot parse app config. Duplicate volume '%s' found in component '%s'.", p, cc.ComponentName)
			}
			mountPoints[p] = p
		}
	}

	// No duplicates detected
	return nil
}

// getAllMountPoints creates a list of all mount points of a component.
func (cc *ComponentConfig) getAllMountPoints(service *ServiceConfig, visitedComponents map[string]string) ([]string, error) {
	// Prevent cycles
	if _, ok := visitedComponents[cc.ComponentName]; ok {
		// Cycle detected
		return nil, errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse app config. Cycle in referenced components detected in '%s'.", cc.ComponentName)
	}
	visitedComponents[cc.ComponentName] = cc.ComponentName

	// Get all mountpoints
	mountPoints := []string{}
	for _, v := range cc.Volumes {
		if v.Path != "" {
			mountPoints = append(mountPoints, v.Path)
		} else if v.VolumePath != "" {
			mountPoints = append(mountPoints, v.VolumePath)
		} else if v.VolumesFrom != "" {
			other := service.findComponent(v.VolumesFrom)
			if other == nil {
				return nil, errgo.WithCausef(nil, InvalidVolumeConfigError, "Cannot parse app config. Cannot find referenced component '%s'.", v.VolumesFrom)
			}
			p, err := other.getAllMountPoints(service, visitedComponents)
			if err != nil {
				return nil, err
			}
			mountPoints = append(mountPoints, p...)
		}
	}
	return mountPoints, nil
}

// validateUniqueDependenciesInPods checks that there are no dependencies with same alias and different port/name
func (sc *ServiceConfig) validateUniqueDependenciesInPods() error {
	// Collect all dependencies per pod
	pod2deps := make(map[string][]DependencyConfig)
	for _, c := range sc.Components {
		pn := c.PodConfig.PodName
		if pn == "" {
			// Not part of a shared pod
			continue
		}
		if c.InstanceConfig.Dependencies == nil {
			// No dependencies
			continue
		}
		list, ok := pod2deps[pn]
		if !ok {
			list = []DependencyConfig{}
		}
		list = append(list, c.InstanceConfig.Dependencies...)
		pod2deps[pn] = list
	}

	// Check each list for duplicates
	for pn, list := range pod2deps {
		for i, dep1 := range list {
			alias1 := dep1.getAlias(sc.ServiceName)
			for j := i + 1; j < len(list); j++ {
				dep2 := list[j]
				alias2 := dep2.getAlias(sc.ServiceName)
				if alias1 == alias2 {
					// Same alias, Port must match and Name must match
					if !dep1.Port.Equals(dep2.Port) {
						return errgo.WithCausef(nil, InvalidDependencyConfigError, "Cannot parse app config. Duplicate (but different ports) dependency '%s' in pod '%s'.", alias1, pn)
					}
					if dep1.Name != dep2.Name {
						return errgo.WithCausef(nil, InvalidDependencyConfigError, "Cannot parse app config. Duplicate (but different names) dependency '%s' in pod '%s'.", alias1, pn)
					}
				}
			}
		}
	}

	// No errors detected
	return nil
}

// validateUniquePortsInPods checks that there are no duplicate ports in a single pod
func (sc *ServiceConfig) validateUniquePortsInPods() error {
	// Collect all ports per pod
	pod2ports := make(map[string][]generictypes.DockerPort)
	for _, c := range sc.Components {
		pn := c.PodConfig.PodName
		if pn == "" {
			// Not part of a shared pod
			continue
		}
		if c.InstanceConfig.Ports == nil {
			// No exposed ports
			continue
		}
		list, ok := pod2ports[pn]
		if !ok {
			list = []generictypes.DockerPort{}
		}
		list = append(list, c.InstanceConfig.Ports...)
		pod2ports[pn] = list
	}

	// Check each list for duplicates
	for pn, list := range pod2ports {
		for i, port1 := range list {
			for j := i + 1; j < len(list); j++ {
				port2 := list[j]
				if port1.Equals(port2) {
					return errgo.WithCausef(nil, InvalidPortConfigError, "Cannot parse app config. Multiple components export port '%s' in pod '%s'.", port1.String(), pn)
				}
			}
		}
	}

	// No errors detected
	return nil
}

// validateScalingPolicyInPods checks that there all scaling policies within a pod are either not set of the same
func (sc *ServiceConfig) validateScalingPolicyInPods() error {
	// Collect all scaling policies per pod
	pod2policies := make(map[string][]ScalingPolicyConfig)
	for _, c := range sc.Components {
		pn := c.PodConfig.PodName
		if pn == "" {
			// Not part of a shared pod
			continue
		}
		if c.ScalingPolicy == nil {
			// No scaling policy set
			continue
		}
		list, ok := pod2policies[pn]
		if !ok {
			list = []ScalingPolicyConfig{}
		}
		list = append(list, *c.ScalingPolicy)
		pod2policies[pn] = list
	}

	// Check each list for errors
	for pn, list := range pod2policies {
		for i, p1 := range list {
			for j := i + 1; j < len(list); j++ {
				p2 := list[j]
				if p1.Min != 0 && p2.Min != 0 {
					// Both minimums specified, must be the same
					if p1.Min != p2.Min {
						return errgo.WithCausef(nil, InvalidScalingConfigError, "Cannot parse app config. Different minimum scaling policies in pod '%s'.", pn)
					}
				}
				if p1.Max != 0 && p2.Max != 0 {
					// Both maximums specified, must be the same
					if p1.Max != p2.Max {
						return errgo.WithCausef(nil, InvalidScalingConfigError, "Cannot parse app config. Different maximum scaling policies in pod '%s'.", pn)
					}
				}
			}
		}
	}

	// No errors detected
	return nil
}

// findComponent finds a component with given name if the list of components inside this service.
// it returns nil if not found
func (sc *ServiceConfig) findComponent(name string) *ComponentConfig {
	for _, c := range sc.Components {
		if c.ComponentName == name {
			return &c
		}
	}
	return nil
}

// getAlias returns the alias of a dependency or its name if there is no alias
func (dc *DependencyConfig) getAlias(serviceName string) string {
	alias := dc.Alias
	if alias == "" {
		_, depComponent := ParseDependency(serviceName, dc.Name)
		alias = depComponent
	}
	return alias
}
