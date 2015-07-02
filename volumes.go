package userconfig

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
			return maskf(InvalidVolumeConfigError, "Cannot parse volume config. Volumes-from for path '%s' should be empty.", vc.Path)
		}
		if vc.VolumeFrom != "" {
			return maskf(InvalidVolumeConfigError, "Cannot parse volume config. Volume-from for path '%s' should be empty.", vc.Path)
		}
		if vc.VolumePath != "" {
			return maskf(InvalidVolumeConfigError, "Cannot parse volume config. Volume-path for path '%s' should be empty.", vc.Path)
		}
		return nil
	}
	// Option 2
	if vc.VolumesFrom != "" {
		if vc.Path != "" {
			return maskf(InvalidVolumeConfigError, "Cannot parse volume config. Path for volumes-from '%s' should be empty.", vc.VolumesFrom)
		}
		if !vc.Size.Empty() {
			return maskf(InvalidVolumeConfigError, "Cannot parse volume config. Size for volumes-from '%s' should be empty.", vc.VolumesFrom)
		}
		if vc.VolumeFrom != "" {
			return maskf(InvalidVolumeConfigError, "Cannot parse volume config. Volume-from for volumes-from '%s' should be empty.", vc.VolumesFrom)
		}
		if vc.VolumePath != "" {
			return maskf(InvalidVolumeConfigError, "Cannot parse volume config. Volume-path for volumes-from '%s' should be empty.", vc.VolumesFrom)
		}
		return nil
	}
	// Option 3
	if vc.VolumeFrom != "" && vc.VolumePath != "" {
		// Path is optional

		if !vc.Size.Empty() {
			return maskf(InvalidVolumeConfigError, "Cannot parse volume config. Size for volume-from '%s' should be empty.", vc.VolumeFrom)
		}
		if vc.VolumesFrom != "" {
			return maskf(InvalidVolumeConfigError, "Cannot parse volume config. Volumes-from for volume-from '%s' should be empty.", vc.VolumeFrom)
		}
		if vc.VolumePath == "" {
			return maskf(InvalidVolumeConfigError, "Cannot parse volume config. Volume-path for volume-from '%s' should not be empty.", vc.VolumeFrom)
		}
		return nil
	}

	// No valid option detected.
	return maskf(InvalidVolumeConfigError, "Cannot parse volume config. Path, volume-path or volumes-path must be set. %#v", vc)
}

func (vc VolumeConfig) V2Validate(valCtx *ValidationContext) error {
	if valCtx == nil {
		return nil
	}

	if vc.Path != "" {
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

type VolumeDefinitions []VolumeConfig

func (vds VolumeDefinitions) validate(valCtx *ValidationContext) error {
	paths := map[string]string{}

	for _, v := range vds {
		if err := v.V2Validate(valCtx); err != nil {
			return mask(err)
		}

		// detect duplicate volume path
		normalized := normalizeFolder(v.Path)
		if _, ok := paths[normalized]; ok {
			return maskf(InvalidVolumeConfigError, "duplicated volume path: %s", normalized)
		}
		paths[normalized] = normalized
	}

	return nil
}
