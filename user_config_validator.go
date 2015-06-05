package userconfig

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

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
		return errgo.WithCausef(nil, ErrUnknownJSONField, "Cannot parse app config. Unknown field '%s' detected.", field[0])
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
func (this *AppDefinition) validate() error {
	for _, s := range this.Services {
		if err := s.validate(); err != nil {
			return err
		}
	}
	if err := this.validateNamespaces(); err != nil {
		return Mask(err)
	}

	return nil
}

// validate performs semantic validations of this ServiceConfig.
// Return the first possible error.
func (this *ServiceConfig) validate() error {
	for _, c := range this.Components {
		if err := c.validate(); err != nil {
			return err
		}
		// Check volume refs
		for _, v := range c.Volumes {
			if err := v.validateRefs(this, &c); err != nil {
				return err
			}
		}
		// Check for duplicate mount points
		if err := c.validateUniqueMountPoints(this); err != nil {
			return err
		}
	}

	return nil
}

// validate performs semantic validations of this ComponentConfig.
// Return the first possible error.
func (this *ComponentConfig) validate() error {
	// Check volumes
	for _, v := range this.Volumes {
		if err := v.validate(); err != nil {
			return err
		}
	}

	for d, _ := range this.Domains {
		if err := d.Validate(); err != nil {
			return Mask(err)
		}
	}

	// No errors found
	return nil
}

type namespaceInfoCounter struct {
	ServiceName string
	Count       int
}

// validateNamespaces checks that
// - namespaces do not cross service boundaries.
// - namespaces must be used in more than 1 component.
func (this *AppDefinition) validateNamespaces() error {
	ns2info := make(map[string]*namespaceInfoCounter)
	for _, s := range this.Services {
		for _, c := range s.Components {
			ns := c.NamespaceName
			if ns != "" {
				info, ok := ns2info[ns]
				if !ok {
					// First occurrence of the namespace
					ns2info[ns] = &namespaceInfoCounter{s.ServiceName, 1}
				} else {
					// Found earlier use of namespace name
					if info.ServiceName != s.ServiceName {
						// Namespace is used in different services
						return errgo.WithCausef(nil, ErrCrossServiceNamespace, "Cannot parse app config. Namespace '%s' is used in multiple services.", ns)
					}
					// Increase counter
					info.Count++
				}
			}
		}
	}
	// Test counters
	for ns, info := range ns2info {
		if info.Count == 1 {
			// Namespace is used only once
			return errgo.WithCausef(nil, ErrNamespaceUsedOnlyOnce, "Cannot parse app config. Namespace '%s' is used in only 1 component.", ns)
		}
	}
	return nil
}

// validate validates the settings of this VolumeConfig.
// Valid combinations:
// - Path & Size set, everything else empty
// - VolumesFrom set, everything else empty
// - VolumeFrom, VolumePath set, Path optionally set, everything else empty
func (this *VolumeConfig) validate() error {
	// Option 1
	if this.Path != "" && !this.Size.Empty() {
		if this.VolumesFrom != "" {
			return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse volume config. Volumes-from for path '%s' should be empty.", this.Path)
		}
		if this.VolumeFrom != "" {
			return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse volume config. Volume-from for path '%s' should be empty.", this.Path)
		}
		if this.VolumePath != "" {
			return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse volume config. Volume-path for path '%s' should be empty.", this.Path)
		}
		return nil
	}
	// Option 2
	if this.VolumesFrom != "" {
		if this.Path != "" {
			return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse volume config. Path for volumes-from '%s' should be empty.", this.VolumesFrom)
		}
		if !this.Size.Empty() {
			return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse volume config. Size for volumes-from '%s' should be empty.", this.VolumesFrom)
		}
		if this.VolumeFrom != "" {
			return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse volume config. Volume-from for volumes-from '%s' should be empty.", this.VolumesFrom)
		}
		if this.VolumePath != "" {
			return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse volume config. Volume-path for volumes-from '%s' should be empty.", this.VolumesFrom)
		}
		return nil
	}
	// Option 3
	if this.VolumeFrom != "" {
		// Path is optional

		if !this.Size.Empty() {
			return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse volume config. Size for volume-from '%s' should be empty.", this.VolumeFrom)
		}
		if this.VolumesFrom != "" {
			return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse volume config. Volumes-from for volume-from '%s' should be empty.", this.VolumeFrom)
		}
		if this.VolumePath == "" {
			return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse volume config. Volume-path for volume-from '%s' should not be empty.", this.VolumeFrom)
		}
		return nil
	}

	// Ok, everything should be empty now
	if this.Path != "" || !this.Size.Empty() || this.VolumePath != "" {
		return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse volume config. Path, volume-path or volumes-path should be set. %#v", this)
	}

	// All empty, ok
	return nil
}

// validateRefs checks the existance of reference names in the given volume config.
func (this *VolumeConfig) validateRefs(service *ServiceConfig, containingComponent *ComponentConfig) error {
	compName := this.VolumesFrom
	if compName == "" {
		compName = this.VolumeFrom
	}
	if compName == "" {
		// No references, all ok
		return nil
	}

	// Check that other component name is not the containing component
	if compName == containingComponent.ComponentName {
		return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse volume config. Cannot refer to own component '%s'.", compName)
	}
	// Another component is referenced, we should be in a namespace
	ns := containingComponent.NamespaceName
	if ns == "" {
		return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse volume config. Cannot refer to another component '%s' without a namespace declaration.", compName)
	}
	// Find the other component name
	other := service.findComponent(compName)
	if other != nil {
		// Found other component
		// Check matching namespace
		if ns != other.NamespaceName {
			return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse volume config. Cannot refer to another component '%s' without a matching namespace declaration.", compName)
		}
		// Check matching "volume-path"
		if this.VolumePath != "" {
			found := false
			for _, v := range other.Volumes {
				if v.Path == this.VolumePath {
					// Found it
					found = true
				}
			}
			if !found {
				return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse volume config. Cannot find path '%s' on component '%s'.", this.VolumePath, compName)
			}
		}
		// all ok
		return nil
	}

	// Not found
	return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse volume config. Cannot find referenced component '%s'.", compName)
}

// validateUniqueMountPoints checks that there are no duplicate volume mounts
func (this *ComponentConfig) validateUniqueMountPoints(service *ServiceConfig) error {
	mountPoints := make(map[string]string)
	for _, v := range this.Volumes {
		var paths []string
		if v.Path != "" {
			paths = []string{v.Path}
		} else if v.VolumeFrom != "" {
			paths = []string{v.VolumePath}
		} else if v.VolumesFrom != "" {
			other := service.findComponent(v.VolumesFrom)
			if other == nil {
				return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse app config. Cannot find referenced component '%s'.", v.VolumesFrom)
			}
			visitedComponents := make(map[string]string)
			var err error
			paths, err = other.getAllMountPoints(service, visitedComponents)
			if err != nil {
				return err
			}
		} else {
			return errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse app config. Missing path in component '%s'.", this.ComponentName)
		}
		for _, p := range paths {
			if _, ok := mountPoints[p]; ok {
				// Found duplicate mount point
				return errgo.WithCausef(nil, ErrDuplicateVolumePath, "Cannot parse app config. Duplicate volume '%s' found in component '%s'.", p, this.ComponentName)
			}
			mountPoints[p] = p
		}
	}

	// No duplicates detected
	return nil
}

// getAllMountPoints creates a list of all mount points of a component.
func (this *ComponentConfig) getAllMountPoints(service *ServiceConfig, visitedComponents map[string]string) ([]string, error) {
	// Prevent cycles
	if _, ok := visitedComponents[this.ComponentName]; ok {
		// Cycle detected
		return nil, errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse app config. Cycle in referenced components detected in '%s'.", this.ComponentName)
	}
	visitedComponents[this.ComponentName] = this.ComponentName

	// Get all mountpoints
	mountPoints := []string{}
	for _, v := range this.Volumes {
		if v.Path != "" {
			mountPoints = append(mountPoints, v.Path)
		} else if v.VolumePath != "" {
			mountPoints = append(mountPoints, v.VolumePath)
		} else if v.VolumesFrom != "" {
			other := service.findComponent(v.VolumesFrom)
			if other == nil {
				return nil, errgo.WithCausef(nil, ErrInvalidVolumeConfig, "Cannot parse app config. Cannot find referenced component '%s'.", v.VolumesFrom)
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

// findComponent finds a component with given name if the list of components inside this service.
// it returns nil if not found
func (this *ServiceConfig) findComponent(name string) *ComponentConfig {
	for _, c := range this.Components {
		if c.ComponentName == name {
			return &c
		}
	}
	return nil
}
