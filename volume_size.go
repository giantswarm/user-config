package userconfig

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

var (
	volumeSizeRegex = regexp.MustCompile("^\\s*([0-9]+)\\s*(GB|G)\\s*$")
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
	size, unit, err := parseVolumeSize(sz)
	if err != nil {
		return mask(err)
	}

	*this = NewVolumeSize(size, unit)
	return nil
}

// Empty returns true if the given volume size in "empty" (unspecified), false otherwise.
func (this VolumeSize) Empty() bool {
	return string(this) == ""
}

// Size gets the size part of a volume size as an integer.
// E.g. "5 GB" -> 5
func (this VolumeSize) Size() (int, error) {
	size, _, err := parseVolumeSize(string(this))
	if err != nil {
		return 0, mask(err)
	}
	return size, nil
}

// Size gets the unit part of a volume size.
// E.g. "5 GB" -> GB
func (this VolumeSize) Unit() (SizeUnit, error) {
	_, unit, err := parseVolumeSize(string(this))
	if err != nil {
		return GB, mask(err)
	}
	return unit, nil
}

// Size gets the size part of a volume size as an integer in GB.
// E.g. "5 GB" -> 5
func (this VolumeSize) SizeInGB() (int, error) {
	if this == "" {
		return 0, nil
	}
	return this.Size()
}

// parseVolumeSize normalizes a string into a well formated VolumeSize.
// if the string cannot be normalized, an error is returned.
func parseVolumeSize(sz string) (int, SizeUnit, error) {
	sz = strings.ToUpper(sz)
	matches := volumeSizeRegex.FindStringSubmatch(sz)
	if matches == nil || len(matches) < 1 || len(matches) > 3 {
		if matches == nil {
			// See if it is just a number
			_, err := strconv.Atoi(sz)
			if err == nil {
				// Just a number, use that without unit for now
				matches = []string{sz, sz}
			}
		}
		// Re-test
		if matches == nil || len(matches) < 1 || len(matches) > 3 {
			return 0, GB, maskf(InvalidSizeError, "Cannot parse app config. Invalid size '%s' detected.", sz)
		}
	}
	unit := "GB"
	if len(matches) > 2 {
		unit = matches[2]
	}
	if unit == "G" {
		unit = "GB"
	}
	// Check size being a proper number
	if nr, err := strconv.ParseUint(matches[1], 10, 32); err != nil {
		return 0, GB, maskf(InvalidSizeError, "Cannot parse app config. Invalid size '%s' detected.", sz)
	} else {
		return int(nr), SizeUnit(unit), nil
	}
}
