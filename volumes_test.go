package userconfig_test

import (
	"encoding/json"
	"testing"

	"github.com/giantswarm/user-config"
)

func TestVolumesInvalidMaxSize(t *testing.T) {
	valCtx := NewValidationContext()

	vd := userconfig.VolumeConfig{Path: "/data", Size: userconfig.VolumeSize("101")}
	err := vd.V2Validate(valCtx)

	if err == nil {
		t.Fatalf("invalid max volume size not detected")
	}
	if !userconfig.IsInvalidVolumeConfig(err) {
		t.Fatalf("expected error to be InvalidVolumeConfigError")
	}
}

func TestVolumesValidMaxSize(t *testing.T) {
	valCtx := NewValidationContext()

	vd := userconfig.VolumeConfig{Path: "/data", Size: userconfig.VolumeSize("100")}
	err := vd.V2Validate(valCtx)

	if err != nil {
		t.Fatalf("volumeDefinition.V2Validate(): %s", err.Error())
	}
}

func TestVolumesDuplicatedPath(t *testing.T) {
	a := ExampleDefinitionWithVolume([]string{"/data", "/data"}, []string{"5 GB", "10 GB"})

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.ServiceDefinition
	err = json.Unmarshal(raw, &b)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %#v", err)
	}
	err = b.Validate(nil)
	if err.Error() != "duplicate volume '/data' found in component 'component/a'" {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsInvalidVolumeConfig(err) {
		t.Fatalf("expected error to be InvalidVolumeConfigError")
	}
}

func TestVolumesDuplicatedPathTrailingSlash(t *testing.T) {
	a := ExampleDefinitionWithVolume([]string{"/data", "/data/"}, []string{"5 GB", "10 GB"})

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.ServiceDefinition
	err = json.Unmarshal(raw, &b)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %#v", err)
	}
	err = b.Validate(nil)
	if err.Error() != "duplicate volume '/data' found in component 'component/a'" {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsInvalidVolumeConfig(err) {
		t.Fatalf("expected error to be InvalidVolumeConfigError")
	}
}

func TestVolumesInvalidSizeUnit(t *testing.T) {
	a := ExampleDefinitionWithVolume([]string{"/data"}, []string{"5 KB"})

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.ServiceDefinition
	err = json.Unmarshal(raw, &b)
	if err == nil {
		t.Fatalf("json.Unmarshal NOT failed")
	}

	if err.Error() != "Cannot parse app config. Invalid size '5 KB' detected." {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsInvalidSize(err) {
		t.Fatalf("expetced error to be InvalidSizeError")
	}
}

func TestVolumesNegativeSize(t *testing.T) {
	a := ExampleDefinitionWithVolume([]string{"/data"}, []string{"-5 GB"})

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.ServiceDefinition
	err = json.Unmarshal(raw, &b)
	if err == nil {
		t.Fatalf("json.Unmarshal NOT failed")
	}

	if err.Error() != "Cannot parse app config. Invalid size '-5 GB' detected." {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsInvalidSize(err) {
		t.Fatalf("expetced error to be InvalidSizeError")
	}
}
