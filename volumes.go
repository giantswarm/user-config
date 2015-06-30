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

func (vd VolumeConfig) V2Validate(valCtx *ValidationContext) error {
	if valCtx == nil {
		return nil
	}

	if vd.Path == "" {
		return maskf(InvalidVolumeConfigError, "volume size cannot be empty")
	}

	intSize, err := vd.Size.SizeInGB()
	if err != nil {
		return maskf(InvalidVolumeConfigError, "invalid volume size '%s', expected '<number> GB'", vd.Size)
	}

	min, err := valCtx.MaxVolumeSize.SizeInGB()
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
