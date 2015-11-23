package userconfig_test

import (
	. "github.com/giantswarm/user-config"
)

func testComponent() *ComponentDefinition {
	return &ComponentDefinition{
		Image: MustParseImageDefinition("registry/namespace/repository:version"),
	}
}

func setPod(config *ComponentDefinition, pod PodEnum) *ComponentDefinition {
	config.Pod = pod
	return config
}

// TODO this is not an app
func testApp() ComponentDefinitions {
	return make(ComponentDefinitions)
}

func testService() V2AppDefinition {
	return V2AppDefinition{
		Components: ComponentDefinitions{},
	}
}
