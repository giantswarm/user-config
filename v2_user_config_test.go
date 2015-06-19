package userconfig_test

import (
	"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
)

func V2ExampleDefinition() userconfig.V2AppDefinition {
	return userconfig.V2AppDefinition{
		Nodes: userconfig.NodeDefinitions{
			userconfig.NodeName("node/a"): userconfig.NodeDefinition{
				Image: generictypes.MustParseDockerImage("registry.giantswarm.io/landingpage:0.10.0"),
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
