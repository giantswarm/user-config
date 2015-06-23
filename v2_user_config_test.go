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
		links = append(links, userconfig.DependencyConfig{Name: name, Port: generictypes.MustParseDockerPort(ports[i])})
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
