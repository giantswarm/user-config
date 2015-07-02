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
	a := V2ExampleDefinitionWithVolume([]string{"/data", "/data"}, []string{"5 GB", "10 GB"})

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.V2AppDefinition
	err = json.Unmarshal(raw, &b)
	if err == nil {
		t.Fatalf("json.Unmarshal NOT failed")
	}

	if err.Error() != "Cannot parse app config. Duplicate volume '/data' found in node 'node/a'." {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsInvalidVolumeConfig(err) {
		t.Fatalf("expetced error to be InvalidVolumeConfigError")
	}
}

func TestVolumesDuplicatedPathTrailingSlash(t *testing.T) {
	a := V2ExampleDefinitionWithVolume([]string{"/data", "/data/"}, []string{"5 GB", "10 GB"})

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.V2AppDefinition
	err = json.Unmarshal(raw, &b)
	if err == nil {
		t.Fatalf("json.Unmarshal NOT failed")
	}

	if err.Error() != "Cannot parse app config. Duplicate volume '/data' found in node 'node/a'." {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsInvalidVolumeConfig(err) {
		t.Fatalf("expetced error to be InvalidVolumeConfigError")
	}
}

func TestVolumesInvalidSizeUnit(t *testing.T) {
	a := V2ExampleDefinitionWithVolume([]string{"/data"}, []string{"5 KB"})

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.V2AppDefinition
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
	a := V2ExampleDefinitionWithVolume([]string{"/data"}, []string{"-5 GB"})

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.V2AppDefinition
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
