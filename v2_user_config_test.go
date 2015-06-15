package userconfig_test

import (
	"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
)

func V2ExampleDefinition() userconfig.V2AppDefinition {
	return userconfig.V2AppDefinition{
		Nodes: map[string]userconfig.NodeDefinition{
			"node/a": userconfig.NodeDefinition{
				Image: generictypes.MustParseDockerImage("registry.giantswarm.io/landingpage:0.10.0"),
				Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
			},
		},
	}
}

func V2ExampleDefinitionWithVolume(path, size string) userconfig.V2AppDefinition {
	appDef := V2ExampleDefinition()
	nodeA, ok := appDef.Nodes["node/a"]
	if !ok {
		panic("missing node")
	}
	nodeA.Volumes = []userconfig.VolumeConfig{
		userconfig.VolumeConfig{Path: path, Size: userconfig.VolumeSize(size)},
	}
	appDef.Nodes["node/a"] = nodeA

	return appDef
}
