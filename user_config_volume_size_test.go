package userconfig_test

import (
	"encoding/json"
	"testing"

	userConfigPkg "github.com/giantswarm/user-config"
)

// Test space + upper case
func TestVolumeSize1(t *testing.T) {
	var compConfig userConfigPkg.ComponentConfig
	byteSlice := []byte(`{
        "component_name": "x",
        "image": "x",
        "volumes": [ { "path": "/tmp", "size": "10 GB" } ]
    }`)

	err := json.Unmarshal(byteSlice, &compConfig)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := string(compConfig.Volumes[0].Size)
	expected := "10 GB"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

// Test space + lower case
func TestVolumeSize2(t *testing.T) {
	var compConfig userConfigPkg.ComponentConfig
	byteSlice := []byte(`{
        "component_name": "x",
        "image": "x",
        "volumes": [ { "path": "/tmp", "size": "10 gb" } ]
    }`)

	err := json.Unmarshal(byteSlice, &compConfig)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := string(compConfig.Volumes[0].Size)
	expected := "10 GB"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

// Test space + mixed case
func TestVolumeSize3(t *testing.T) {
	var compConfig userConfigPkg.ComponentConfig
	byteSlice := []byte(`{
        "component_name": "x",
        "image": "x",
        "volumes": [ { "path": "/tmp", "size": "10 Gb" } ]
    }`)

	err := json.Unmarshal(byteSlice, &compConfig)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := string(compConfig.Volumes[0].Size)
	expected := "10 GB"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

// Test multiple spaces + upper case
func TestVolumeSize1a(t *testing.T) {
	var compConfig userConfigPkg.ComponentConfig
	byteSlice := []byte(`{
        "component_name": "x",
        "image": "x",
        "volumes": [ { "path": "/tmp", "size": "10  GB" } ]
    }`)

	err := json.Unmarshal(byteSlice, &compConfig)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := string(compConfig.Volumes[0].Size)
	expected := "10 GB"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

// Test multiple spaces + lower case
func TestVolumeSize2a(t *testing.T) {
	var compConfig userConfigPkg.ComponentConfig
	byteSlice := []byte(`{
        "component_name": "x",
        "image": "x",
        "volumes": [ { "path": "/tmp", "size": "10  gb" } ]
    }`)

	err := json.Unmarshal(byteSlice, &compConfig)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := string(compConfig.Volumes[0].Size)
	expected := "10 GB"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

// Test multiple spaces + mixed case
func TestVolumeSize3a(t *testing.T) {
	var compConfig userConfigPkg.ComponentConfig
	byteSlice := []byte(`{
        "component_name": "x",
        "image": "x",
        "volumes": [ { "path": "/tmp", "size": "10  Gb" } ]
    }`)

	err := json.Unmarshal(byteSlice, &compConfig)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := string(compConfig.Volumes[0].Size)
	expected := "10 GB"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

// Test no space + upper case
func TestVolumeSize1b(t *testing.T) {
	var compConfig userConfigPkg.ComponentConfig
	byteSlice := []byte(`{
        "component_name": "x",
        "image": "x",
        "volumes": [ { "path": "/tmp", "size": "34GB" } ]
    }`)

	err := json.Unmarshal(byteSlice, &compConfig)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := string(compConfig.Volumes[0].Size)
	expected := "34 GB"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

// Test no space + lower case
func TestVolumeSize2b(t *testing.T) {
	var compConfig userConfigPkg.ComponentConfig
	byteSlice := []byte(`{
        "component_name": "x",
        "image": "x",
        "volumes": [ { "path": "/tmp", "size": "1gb" } ]
    }`)

	err := json.Unmarshal(byteSlice, &compConfig)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := string(compConfig.Volumes[0].Size)
	expected := "1 GB"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

// Test no space + mixed case
func TestVolumeSize3b(t *testing.T) {
	var compConfig userConfigPkg.ComponentConfig
	byteSlice := []byte(`{
        "component_name": "x",
        "image": "x",
        "volumes": [ { "path": "/tmp", "size": "100000 Gb" } ]
    }`)

	err := json.Unmarshal(byteSlice, &compConfig)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := string(compConfig.Volumes[0].Size)
	expected := "100000 GB"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

// Test no space + upper case + "G"
func TestVolumeSize1c(t *testing.T) {
	var compConfig userConfigPkg.ComponentConfig
	byteSlice := []byte(`{
        "component_name": "x",
        "image": "x",
        "volumes": [ { "path": "/tmp", "size": "34G" } ]
    }`)

	err := json.Unmarshal(byteSlice, &compConfig)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := string(compConfig.Volumes[0].Size)
	expected := "34 GB"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

// Test no space + lower case + "g"
func TestVolumeSize2c(t *testing.T) {
	var compConfig userConfigPkg.ComponentConfig
	byteSlice := []byte(`{
        "component_name": "x",
        "image": "x",
        "volumes": [ { "path": "/tmp", "size": "1g" } ]
    }`)

	err := json.Unmarshal(byteSlice, &compConfig)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := string(compConfig.Volumes[0].Size)
	expected := "1 GB"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

// Test no space + mixed case + "G"
func TestVolumeSize3c(t *testing.T) {
	var compConfig userConfigPkg.ComponentConfig
	byteSlice := []byte(`{
        "component_name": "x",
        "image": "x",
        "volumes": [ { "path": "/tmp", "size": "100000 G" } ]
    }`)

	err := json.Unmarshal(byteSlice, &compConfig)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := string(compConfig.Volumes[0].Size)
	expected := "100000 GB"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

func TestVolumeSizeGetters(t *testing.T) {
	tests := []struct {
		Input           string
		ExpectSizeError bool
		ExpectedSize    int
		ExpectUnitError bool
		ExpectedUnit    userConfigPkg.SizeUnit
	}{
		{"5 GB", false, 5, false, userConfigPkg.GB},
		{"123 G", false, 123, false, userConfigPkg.GB},
		{"123", false, 123, true, userConfigPkg.GB},
		{"abc G", true, 0, false, userConfigPkg.GB},
		{"", true, 0, true, userConfigPkg.GB},
		{"124 KB", false, 124, true, userConfigPkg.GB},
		{"5GB", true, 0, true, userConfigPkg.GB},
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
	}
}
