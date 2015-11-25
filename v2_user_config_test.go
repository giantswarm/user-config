package userconfig_test

import (
	"encoding/json"
	"testing"

	"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
)

func ExampleDefinition() userconfig.ServiceDefinition {
	return userconfig.ServiceDefinition{
		Components: userconfig.ComponentDefinitions{
			userconfig.ComponentName("component/a"): &userconfig.ComponentDefinition{
				Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
				Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
			},
			userconfig.ComponentName("component/b"): &userconfig.ComponentDefinition{
				Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/giantswarm/b:0.10.0"),
				Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
			},
		},
	}
}

func ExampleDefinitionWithVolume(paths, sizes []string) userconfig.ServiceDefinition {
	serviceDef := ExampleDefinition()
	componentA, ok := serviceDef.Components["component/a"]
	if !ok {
		panic("missing component")
	}

	if len(paths) != len(sizes) {
		panic("list of path and size must be equal")
	}
	volumes := userconfig.VolumeDefinitions{}
	for i, path := range paths {
		volumes = append(volumes, userconfig.VolumeConfig{Path: path, Size: userconfig.VolumeSize(sizes[i])})
	}
	componentA.Volumes = volumes
	serviceDef.Components["component/a"] = componentA

	return serviceDef
}

func ExampleDefinitionWithLinks(names, ports []string) userconfig.ServiceDefinition {
	serviceDef := ExampleDefinition()
	componentA, ok := serviceDef.Components["component/a"]
	if !ok {
		panic("missing component")
	}

	if len(names) != len(ports) {
		panic("list of names and ports must be equal")
	}
	links := userconfig.LinkDefinitions{}
	for i, name := range names {
		links = append(links, userconfig.LinkDefinition{Component: userconfig.ComponentName(name), TargetPort: generictypes.MustParseDockerPort(ports[i])})
	}
	componentA.Links = links
	serviceDef.Components["component/a"] = componentA

	return serviceDef
}

func NewValidationContext() *userconfig.ValidationContext {
	return &userconfig.ValidationContext{
		Protocols:     []string{generictypes.ProtocolTCP},
		MinScaleSize:  1,
		MaxScaleSize:  10,
		MinVolumeSize: userconfig.NewVolumeSize(1, userconfig.GB),
		MaxVolumeSize: userconfig.NewVolumeSize(100, userconfig.GB),
	}
}

func TestServiceValidLinks(t *testing.T) {
	a := ExampleDefinitionWithLinks([]string{"component/b"}, []string{"80/tcp"})
	_, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
}

func TestServiceLinksInvalidComponent(t *testing.T) {
	a := ExampleDefinitionWithLinks([]string{"component/c"}, []string{"80/tcp"})
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
	if err.Error() != "invalid link to component 'component/c': does not exists" {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsInvalidComponentDefinition(err) {
		t.Fatalf("expetced error to be InvalidComponentDefinitionError")
	}
}

// That test is usefull to ensure that `swarm cat` works as expected. There was
// an issue where the service def was marshaled and unmarshaled twice on its way
// from appd to api to cli. There the scale was defaulted although none was set
// by the user. This was caused by a wrong implementation in the service def
// validation.
func TestServiceMarshalUnmarshalDontSetDefaults(t *testing.T) {
	a := ExampleDefinition()

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.ServiceDefinition
	err = json.Unmarshal(raw, &b)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %s", err.Error())
	}

	if b.Components["component/a"].Scale != nil {
		t.Fatalf("scale not hidden")
	}
}

func TestServiceSetDefaults(t *testing.T) {
	a := ExampleDefinition()
	valCtx := NewValidationContext()

	if err := a.SetDefaults(valCtx); err != nil {
		t.Fatalf("setting defaults failed: %#v", err)
	}

	if err := a.Validate(valCtx); err != nil {
		t.Fatalf("validating service failed: %#v", err)
	}

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.ServiceDefinition
	err = json.Unmarshal(raw, &b)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %s", err.Error())
	}

	if b.Components["component/a"].Scale.Min != valCtx.MinScaleSize {
		t.Fatalf("min scale size not set")
	}

	if b.Components["component/a"].Scale.Max != valCtx.MaxScaleSize {
		t.Fatalf("max scale size not set")
	}
}

func TestServiceHideDefaults(t *testing.T) {
	a := ExampleDefinition()
	valCtx := NewValidationContext()

	if err := a.SetDefaults(valCtx); err != nil {
		t.Fatalf("setting defaults failed: %#v", err)
	}

	if err := a.Validate(valCtx); err != nil {
		t.Fatalf("validating service failed: %#v", err)
	}

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.ServiceDefinition
	err = json.Unmarshal(raw, &b)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %s", err.Error())
	}

	c, err := b.HideDefaults(valCtx)
	if err != nil {
		t.Fatalf("hiding defaults failed: %s", err.Error())
	}

	if c.Components["component/a"].Scale != nil {
		t.Fatalf("scale not hidden")
	}
}

func TestAbsentServiceName(t *testing.T) {
	a := ExampleDefinition()
	name, err := a.Name()
	if err != nil {
		t.Fatalf("Name failed: %#v", err)
	}
	expectedName := "2308909c"
	if name != expectedName {
		t.Fatalf("Name result is invalid, got '%s', expected '%s'", name, expectedName)
	}
}

func TestSpecifiedServiceName(t *testing.T) {
	a := ExampleDefinition()
	expectedName := "nice-he"
	a.ServiceName = userconfig.ServiceName(expectedName)
	name, err := a.Name()
	if err != nil {
		t.Fatalf("Name failed: %#v", err)
	}
	if name != expectedName {
		t.Fatalf("Name result is invalid, got '%s', expected '%s'", name, expectedName)
	}
}
