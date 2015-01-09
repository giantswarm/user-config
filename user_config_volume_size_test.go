package userconfig_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/juju/errgo"

	userConfigPkg "github.com/giantswarm/user-config"
)

// Various volume size unmarshal tests
func TestVolumeSizeUnmarshal(t *testing.T) {
	tests := []struct {
		Input       string
		ExpectedErr error
		Result      string
	}{
		// Correct ones
		{"1 G", nil, "1 GB"},
		{"10 GB", nil, "10 GB"},
		{"5G", nil, "5 GB"},
		{"8GB", nil, "8 GB"},
		{"100", nil, "100 GB"},
		// Correct, mixed case
		{"8gb", nil, "8 GB"},
		{"8gB", nil, "8 GB"},
		{"8Gb", nil, "8 GB"},
		{"9 gb", nil, "9 GB"},
		{"9 gB", nil, "9 GB"},
		{"9 Gb", nil, "9 GB"},
		// Correct, extra spaces
		{"   8gb", nil, "8 GB"},
		{"  8gB     ", nil, "8 GB"},
		{"8Gb   ", nil, "8 GB"},
		{"8    Gb   ", nil, "8 GB"},

		// Invalid ones
		{"-0 G", userConfigPkg.ErrInvalidSize, ""},
		{"-9223372036854775806 G", userConfigPkg.ErrInvalidSize, ""},
		{"- 9223372036854775806 G", userConfigPkg.ErrInvalidSize, ""},
		{"-0 GB", userConfigPkg.ErrInvalidSize, ""},
		{"-9223372036854775806 GB", userConfigPkg.ErrInvalidSize, ""},
		{"- 9223372036854775806 GB", userConfigPkg.ErrInvalidSize, ""},
		{"-0G", userConfigPkg.ErrInvalidSize, ""},
		{"-9223372036854775806G", userConfigPkg.ErrInvalidSize, ""},
		{"- 9223372036854775806G", userConfigPkg.ErrInvalidSize, ""},
		{"-0GB", userConfigPkg.ErrInvalidSize, ""},
		{"-9223372036854775806GB", userConfigPkg.ErrInvalidSize, ""},
		{"- 9223372036854775806GB", userConfigPkg.ErrInvalidSize, ""},
		{"-9223372036854775806", userConfigPkg.ErrInvalidSize, ""},
		{"GB 10", userConfigPkg.ErrInvalidSize, ""},
		{" 0x10 GB", userConfigPkg.ErrInvalidSize, ""},
	}

	for _, test := range tests {
		var compConfig userConfigPkg.ComponentConfig
		byteSlice := []byte(fmt.Sprintf(`{
        "component_name": "x",
        "image": "x",
        "volumes": [ { "path": "/tmp", "size": "%s" } ]
    }`, test.Input))

		err := json.Unmarshal(byteSlice, &compConfig)
		if err != nil {
			if errgo.Cause(err) != test.ExpectedErr {
				t.Fatalf("Unmarshal failed: %v", err)
			}
		} else {
			got := string(compConfig.Volumes[0].Size)
			expected := test.Result
			if got != expected {
				t.Fatalf("Invalid result: got %s, expected %s", got, expected)
			}
		}
	}
}

func TestVolumeSizeGetters(t *testing.T) {
	tests := []struct {
		Input                 string
		ExpectSizeError       bool
		ExpectedSize          int
		ExpectUnitError       bool
		ExpectedUnit          userConfigPkg.SizeUnit
		ExpectedSizeInGBError bool
		ExpectedSizeInGB      int
	}{
		{"5 GB", false, 5, false, userConfigPkg.GB, false, 5},
		{"123 G", false, 123, false, userConfigPkg.GB, false, 123},
		{"123", false, 123, true, userConfigPkg.GB, true, 123},
		{"abc G", true, 0, false, userConfigPkg.GB, true, 0},
		{"", true, 0, true, userConfigPkg.GB, false, 0},
		{"124 KB", false, 124, true, userConfigPkg.GB, true, 124},
		{"5GB", true, 0, true, userConfigPkg.GB, true, 0},
		{"8", false, 8, false, userConfigPkg.GB, false, 8},
	}

	for _, test := range tests {
		input := userConfigPkg.VolumeSize(test.Input)
		if sz, err := input.Size(); err != nil {
			if !test.ExpectSizeError {
				t.Fatalf("Size failed: %v", err)
			}
		} else if sz != test.ExpectedSize {
			t.Fatalf("Invalid Size() result: got %v, expected %v", sz, test.ExpectedSize)
		}
		if unit, err := input.Unit(); err != nil {
			if !test.ExpectUnitError {
				t.Fatalf("Unit failed: %v", err)
			}
		} else if unit != test.ExpectedUnit {
			t.Fatalf("Invalid Unit() result: got %v, expected %v", unit, test.ExpectedUnit)
		}
		if sz, err := input.SizeInGB(); err != nil {
			if !test.ExpectedSizeInGBError {
				t.Fatalf("SizeInGB failed: %v", err)
			}
		} else if sz != test.ExpectedSizeInGB {
			t.Fatalf("Invalid SizeInGB() result: got %v, expected %v", sz, test.ExpectedSizeInGB)
		}
	}
}

func TestVolumeSizeNew(t *testing.T) {
	tests := []struct {
		Size     int
		Unit     userConfigPkg.SizeUnit
		Expected string
	}{
		{5, userConfigPkg.GB, "5 GB"},
		{110, userConfigPkg.GB, "110 GB"},
	}

	for _, test := range tests {
		vs := userConfigPkg.NewVolumeSize(test.Size, test.Unit)
		if string(vs) != test.Expected {
			t.Fatalf("Invalid result: got %v, expected %v", string(vs), test.Expected)
		}
	}
}
