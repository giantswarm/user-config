package userconfig

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/juju/errgo"
)

var (
	volumeSizeRegex = regexp.MustCompile("([0-9]+)\\s*(GB|G)")
)

type VolumeSize string
type SizeUnit string

const (
	GB = SizeUnit("GB")
)

// NewVolumeSize creates a new VolumeSize with given parameters
func NewVolumeSize(size int, unit SizeUnit) VolumeSize {
	sz := strconv.Itoa(size)
	return VolumeSize(sz + " " + string(unit))
}

// UnmarshalJSON performs a format friendly parsing of volume sizes
func (this *VolumeSize) UnmarshalJSON(data []byte) error {
	var sz string
	err := json.Unmarshal(data, &sz)
	if err != nil {
		return err
	}
	sz = strings.ToUpper(sz)
	matches := volumeSizeRegex.FindStringSubmatch(sz)
	if matches == nil || len(matches) != 3 {
		return errgo.WithCausef(nil, ErrInvalidSize, "Cannot parse app config. Invalid size '%s' detected.", sz)
	}
	unit := matches[2]
	if unit == "G" {
		unit = "GB"
	}
	*this = VolumeSize(matches[1] + " " + unit)
	return nil
}

// Size gets the size part of a volume size as an integer.
// E.g. "5 GB" -> 5
func (this VolumeSize) Size() (int, error) {
	parts := strings.Split(string(this), " ")
	return strconv.Atoi(parts[0])
}

// Size gets the unit part of a volume size.
// E.g. "5 GB" -> GB
func (this VolumeSize) Unit() (SizeUnit, error) {
	parts := strings.Split(string(this), " ")
	if len(parts) < 2 {
		return GB, errgo.Newf("No unit found, got '%s'", string(this))
	}
	switch parts[1] {
	case "G":
		return GB, nil
	case "GB":
		return GB, nil
	default:
		return GB, errgo.Newf("Unknown unit, got '%s'", parts[1])
	}
}

// Size gets the size part of a volume size as an integer in GB.
// E.g. "5 GB" -> 5
func (this VolumeSize) SizeInGB() (int, error) {
	if this == "" {
		return 0, nil
	}
	return this.Size()
}
