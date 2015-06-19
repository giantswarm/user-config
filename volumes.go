package userconfig

import (
	"github.com/juju/errgo"
)

var (
	defaultMinVolumeSize = 1
	MinVolumeSize        = 0
	defaultMaxVolumeSize = 100
	MaxVolumeSize        = 0
)

func init() {
	if MinVolumeSize == 0 {
		MinVolumeSize = defaultMinVolumeSize
	}
	if MaxVolumeSize == 0 {
		MaxVolumeSize = defaultMaxVolumeSize
	}
}

func SetMinVolumeSize(min int) {
	MinVolumeSize = min
}

func SetMaxVolumeSize(max int) {
	MaxVolumeSize = max
}

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

func (vd VolumeConfig) v2Validate() error {
	if vd.Path == "" {
		return Mask(errgo.WithCausef(nil, InvalidVolumeConfigError, "volume size cannot be empty"))
	}

	intSize, err := vd.Size.SizeInGB()
	if err != nil {
		return Mask(errgo.WithCausef(nil, InvalidVolumeConfigError, "invalid volume size '%s', expected '<number> GB'", vd.Size))
	}

	if intSize < MinVolumeSize {
		return Mask(errgo.WithCausef(nil, InvalidVolumeConfigError, "volume size '%d' cannot be less than '%d'", intSize, MinVolumeSize))
	}

	if intSize > MaxVolumeSize {
		return Mask(errgo.WithCausef(nil, InvalidVolumeConfigError, "volume size '%d' cannot be greater than '%d'", intSize, MaxVolumeSize))
	}

	return nil
}

type VolumeDefinitions []VolumeConfig

func (vds VolumeDefinitions) validate() error {
	paths := map[string]string{}

	for _, v := range vds {
		if err := v.v2Validate(); err != nil {
			return Mask(err)
		}

		// detect duplicate volume path
		if _, ok := paths[v.Path]; ok {
			return Mask(errgo.WithCausef(nil, InvalidVolumeConfigError, "duplicated volume path: %s", v.Path))
		}
		paths[v.Path] = v.Path
	}

	return nil
}
