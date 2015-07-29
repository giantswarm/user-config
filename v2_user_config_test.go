package userconfig_test

import (
	"encoding/json"
	"testing"

	"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
)

func V2ExampleDefinition() userconfig.V2AppDefinition {
	return userconfig.V2AppDefinition{
		Nodes: userconfig.NodeDefinitions{
			userconfig.NodeName("node/a"): &userconfig.NodeDefinition{
				Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
				Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
			},
			userconfig.NodeName("node/b"): &userconfig.NodeDefinition{
				Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/giantswarm/b:0.10.0"),
				Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
			},
		},
	}
}

func V2ExampleDefinitionWithVolume(paths, sizes []string) userconfig.V2AppDefinition {
	appDef := V2ExampleDefinition()
	nodeA, ok := appDef.Nodes["node/a"]
	if !ok {
		panic("missing node")
	}

	if len(paths) != len(sizes) {
		panic("list of path and size must be equal")
	}
	volumes := userconfig.VolumeDefinitions{}
	for i, path := range paths {
		volumes = append(volumes, userconfig.VolumeConfig{Path: path, Size: userconfig.VolumeSize(sizes[i])})
	}
	nodeA.Volumes = volumes
	appDef.Nodes["node/a"] = nodeA

	return appDef
}

func V2ExampleDefinitionWithLinks(names, ports []string) userconfig.V2AppDefinition {
	appDef := V2ExampleDefinition()
	nodeA, ok := appDef.Nodes["node/a"]
	if !ok {
		panic("missing node")
	}

	if len(names) != len(ports) {
		panic("list of names and ports must be equal")
	}
	links := userconfig.LinkDefinitions{}
	for i, name := range names {
		links = append(links, userconfig.LinkDefinition{Name: userconfig.NodeName(name), Port: generictypes.MustParseDockerPort(ports[i])})
	}
	nodeA.Links = links
	appDef.Nodes["node/a"] = nodeA

	return appDef
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

func TestV2AppValidLinks(t *testing.T) {
	a := V2ExampleDefinitionWithLinks([]string{"node/b"}, []string{"80/tcp"})
	_, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
}

func TestV2AppLinksInvalidNode(t *testing.T) {
	a := V2ExampleDefinitionWithLinks([]string{"node/c"}, []string{"80/tcp"})
	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.V2AppDefinition
	err = json.Unmarshal(raw, &b)
	if err == nil {
		t.Fatalf("json.Unmarshal NOT failed")
	}
	if err.Error() != "invalid link to node 'node/c': does not exists" {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsInvalidNodeDefinition(err) {
		t.Fatalf("expetced error to be InvalidNodeDefinitionError")
	}
}

// That test is usefull to ensure that `swarm cat` works as expected. There was
// an issue where the app def was marshaled and unmarshaled twice on its way
// from appd to api to cli. There the scale was defaulted although none was set
// by the user. This was caused by a wrong implementation in the app def
// validation.
func TestV2AppMarshalUnmarshalDontSetDefaults(t *testing.T) {
	a := V2ExampleDefinition()

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.V2AppDefinition
	err = json.Unmarshal(raw, &b)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %s", err.Error())
	}

	if b.Nodes["node/a"].Scale != nil {
		t.Fatalf("scale not hidden")
	}
}

func TestV2AppSetDefaults(t *testing.T) {
	a := V2ExampleDefinition()
	valCtx := NewValidationContext()

	if err := a.SetDefaults(valCtx); err != nil {
		t.Fatalf("setting defaults failed: %#v", err)
	}

	if err := a.Validate(valCtx); err != nil {
		t.Fatalf("validating app failed: %#v", err)
	}

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.V2AppDefinition
	err = json.Unmarshal(raw, &b)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %s", err.Error())
	}

	if b.Nodes["node/a"].Scale.Min != valCtx.MinScaleSize {
		t.Fatalf("min scale size not set")
	}

	if b.Nodes["node/a"].Scale.Max != valCtx.MaxScaleSize {
		t.Fatalf("max scale size not set")
	}
}

func TestV2AppHideDefaults(t *testing.T) {
	a := V2ExampleDefinition()
	valCtx := NewValidationContext()

	if err := a.SetDefaults(valCtx); err != nil {
		t.Fatalf("setting defaults failed: %#v", err)
	}

	if err := a.Validate(valCtx); err != nil {
		t.Fatalf("validating app failed: %#v", err)
	}

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.V2AppDefinition
	err = json.Unmarshal(raw, &b)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %s", err.Error())
	}

	c, err := b.HideDefaults(valCtx)
	if err != nil {
		t.Fatalf("hiding defaults failed: %s", err.Error())
	}

	if c.Nodes["node/a"].Scale != nil {
		t.Fatalf("scale not hidden")
	}
}

func TestV2AbsentAppName(t *testing.T) {
	a := V2ExampleDefinition()
	name, err := a.Name()
	if err != nil {
		t.Fatalf("Name failed: %#v", err)
	}
	expectedName := "e27445c0"
	if name != expectedName {
		t.Fatalf("Name result is invalid, got '%s', expected '%s'", name, expectedName)
	}
}

func TestV2SpecifiedAppName(t *testing.T) {
	a := V2ExampleDefinition()
	expectedName := "nice-he"
	a.AppName = userconfig.AppName(expectedName)
	name, err := a.Name()
	if err != nil {
		t.Fatalf("Name failed: %#v", err)
	}
	if name != expectedName {
		t.Fatalf("Name result is invalid, got '%s', expected '%s'", name, expectedName)
	}
}
