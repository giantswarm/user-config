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
