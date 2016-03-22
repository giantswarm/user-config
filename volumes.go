package userconfig

import (
	"encoding/json"
	"fmt"
	"sort"
)

type VolumeConfig struct {
	// Path of the volume to mount, e.g. "/opt/service/".
	Path string `json:"path,omitempty" description:"Path of the volume to mount (inside the container)`

	// Storage size in GB, e.g. "5 GB".
	Size VolumeSize `json:"size,omitempty" description:"Size of the volume. e.g. '5 GB'"`

	// Name of another component to map all volumes from
	VolumesFrom string `json:"volumes-from,omitempty" description:"Name of another component (in same pod) to share volumes with"`

	// Name of another component to map a specific volumes from
	VolumeFrom string `json:"volume-from,omitempty" description:"Name of another component (in same pod) to share a specific volume with"`

	// Path inside the other component to share
	VolumePath string `json:"volume-path,omitempty" description:"Path in another component to share"`

	// Shared controls whether all instances of the parent component will get the same volume (true) or each gets their own (false). Defaults to false.
	Shared bool `json:"shared,omitempty" description:"Share this volume across all instances"`
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
			return maskf(InvalidVolumeConfigError, "volumes-from for path '%s' should be empty", vc.Path)
		}
		if vc.VolumeFrom != "" {
			return maskf(InvalidVolumeConfigError, "volume-from for path '%s' should be empty", vc.Path)
		}
		if vc.VolumePath != "" {
			return maskf(InvalidVolumeConfigError, "volume-path for path '%s' should be empty", vc.Path)
		}
		return nil
	}
	// Option 2
	if vc.VolumesFrom != "" {
		if vc.Path != "" {
			return maskf(InvalidVolumeConfigError, "path for volumes-from '%s' should be empty", vc.VolumesFrom)
		}
		if !vc.Size.Empty() {
			return maskf(InvalidVolumeConfigError, "size for volumes-from '%s' should be empty", vc.VolumesFrom)
		}
		if vc.VolumeFrom != "" {
			return maskf(InvalidVolumeConfigError, "volume-from for volumes-from '%s' should be empty", vc.VolumesFrom)
		}
		if vc.VolumePath != "" {
			return maskf(InvalidVolumeConfigError, "volume-path for volumes-from '%s' should be empty", vc.VolumesFrom)
		}
		return nil
	}
	// Option 3
	if vc.VolumeFrom != "" && vc.VolumePath != "" {
		// Path is optional

		if !vc.Size.Empty() {
			return maskf(InvalidVolumeConfigError, "size for volume-from '%s' should be empty", vc.VolumeFrom)
		}
		if vc.VolumesFrom != "" {
			return maskf(InvalidVolumeConfigError, "volumes-from for volume-from '%s' should be empty", vc.VolumeFrom)
		}
		if vc.VolumePath == "" {
			return maskf(InvalidVolumeConfigError, "volume-path for volume-from '%s' should not be empty", vc.VolumeFrom)
		}
		return nil
	}

	// No valid option detected.
	return maskf(InvalidVolumeConfigError, "path & size, volume-path or volumes-path must be set in '%#v'", vc)
}

func (vc VolumeConfig) V2Validate(valCtx *ValidationContext) error {
	if valCtx == nil {
		return nil
	}

	if vc.Path != "" && vc.Size != "" {
		intSize, err := vc.Size.SizeInGB()
		if err != nil {
			return maskf(InvalidVolumeConfigError, "invalid volume size '%s', expected '<number> GB'", vc.Size)
		}

		min, err := valCtx.MinVolumeSize.SizeInGB()
		if err != nil {
			return mask(err)
		}

		if intSize < min {
			return maskf(InvalidVolumeConfigError, "volume size '%d' cannot be less than '%d'", intSize, min)
		}

		max, err := valCtx.MaxVolumeSize.SizeInGB()
		if err != nil {
			return mask(err)
		}

		if intSize > max {
			return maskf(InvalidVolumeConfigError, "volume size '%d' cannot be greater than '%d'", intSize, max)
		}
	}

	// Check other properties
	if err := vc.validate(); err != nil {
		return mask(err)
	}

	return nil
}

func (vd VolumeConfig) String() string {
	raw, err := json.Marshal(vd)
	if err != nil {
		panic(fmt.Sprintf("%#v\n", mask(err)))
	}

	return string(raw)
}

type VolumeDefinitions []VolumeConfig

// String returns the marshalled and ordered string represantion of its own
// incarnation. It is important to have the string represantion ordered, since
// we use it to compare two VolumeDefinitions when creating a diff. See diff.go
func (vds VolumeDefinitions) String() string {
	list := []string{}

	for _, vd := range vds {
		list = append(list, vd.String())
	}
	sort.Strings(list)

	raw, err := json.Marshal(list)
	if err != nil {
		panic(fmt.Sprintf("%#v\n", mask(err)))
	}

	return string(raw)
}

func (vds VolumeDefinitions) Empty() bool {
	return len(vds) == 0
}

// Contains returns true if the volumes contain a volume with the given path,
// or false otherwise.
func (vds VolumeDefinitions) Contains(path string) bool {
	for _, v := range vds {
		if v.Path == path {
			return true
		}
	}
	return false
}

func (vds VolumeDefinitions) validate(valCtx *ValidationContext) error {
	for _, v := range vds {
		if err := v.V2Validate(valCtx); err != nil {
			return mask(err)
		}
	}

	return nil
}

// validateVolumesRefs checks for each volume in each component the existance of reference names in the given volume config.
func (nds *ComponentDefinitions) validateVolumesRefs() error {
	for componentName, componentDef := range *nds {
		for _, vc := range componentDef.Volumes {
			if err := nds.validateVolumeRefs(vc, componentName); err != nil {
				return mask(err)
			}
		}
	}
	return nil
}

// validateVolumeRefs checks the existance of reference names in the given volume config.
func (nds *ComponentDefinitions) validateVolumeRefs(vc VolumeConfig, containingComponentName ComponentName) error {
	componentName := vc.VolumesFrom
	if componentName == "" {
		componentName = vc.VolumeFrom
	}
	if componentName == "" {
		// No references, all ok
		return nil
	}

	// Check that the component name (volume-from or volumes-from) is not the containing component
	if componentName == containingComponentName.String() {
		return maskf(InvalidVolumeConfigError, "cannot refer to own component '%s'", componentName)
	}
	// Another component is referenced, we should be in a pod
	// Find the root of our pod
	podRootName, _, err := nds.PodRoot(containingComponentName)
	if err != nil {
		return maskf(InvalidVolumeConfigError, "cannot refer to another component '%s' without a pod declaration", componentName)
	}
	// Get the components that are part of the same pod
	podComponents, err := nds.PodComponents(podRootName)
	if err != nil {
		return mask(err)
	}
	// Find the other component name
	other, err := podComponents.ComponentByName(ComponentName(componentName))
	if err == nil {
		// Found other component
		// Check matching "volume-path"
		if vc.VolumePath != "" {
			if !other.Volumes.Contains(vc.VolumePath) {
				return maskf(InvalidVolumeConfigError, "cannot find path '%s' on component '%s'", vc.VolumePath, componentName)
			}
		}
		// all ok
		return nil
	}

	// Other component is not found in the same pod
	// Does the other component even exists?
	if _, err := nds.ComponentByName(ComponentName(componentName)); err == nil {
		return maskf(InvalidVolumeConfigError, "cannot refer to another component '%s' that is not part of the same pod", componentName)
	} else {
		// Other component not found
		return maskf(InvalidVolumeConfigError, "cannot find referenced component '%s'", componentName)
	}
}

// validateUniqueMountPoints checks that there are no duplicate volume mounts
func (nds *ComponentDefinitions) validateUniqueMountPoints() error {
	for componentName, componentDef := range *nds {
		mountPoints := make(map[string]string)
		for _, v := range componentDef.Volumes {
			var paths []string
			if v.Path != "" {
				paths = []string{normalizeFolder(v.Path)}
			} else if v.VolumeFrom != "" {
				paths = []string{normalizeFolder(v.VolumePath)}
			} else if v.VolumesFrom != "" {
				if _, err := nds.ComponentByName(ComponentName(v.VolumesFrom)); err != nil {
					return maskf(InvalidVolumeConfigError, "cannot find referenced component '%s'", v.VolumesFrom)
				}
				var err error
				paths, err = nds.MountPoints(ComponentName(v.VolumesFrom))
				if err != nil {
					return mask(err)
				}
			} else {
				return maskf(InvalidVolumeConfigError, "missing path in component '%s'", componentName.String())
			}
			for _, p := range paths {
				if _, ok := mountPoints[p]; ok {
					// Found duplicate mount point
					return maskf(DuplicateVolumePathError, "duplicate volume '%s' found in component '%s'", p, componentName.String())
				}
				mountPoints[p] = p
			}
		}
	}

	// No duplicates detected
	return nil
}
