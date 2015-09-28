package userconfig_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/giantswarm/generic-types-go"
	. "github.com/giantswarm/user-config"
)

func testDiffCallWith(t *testing.T, oldDef, newDef V2AppDefinition, expectedDiffInfos []DiffInfo) {
	diffInfos := ServiceDiff(oldDef, newDef)

	if len(diffInfos) != len(expectedDiffInfos) {
		t.Fatalf("Expected %d item, got %d: %#v", len(expectedDiffInfos), len(diffInfos), diffInfos)
	}

	if !reflect.DeepEqual(diffInfos, expectedDiffInfos) {
		for _, exp := range expectedDiffInfos {
			t.Logf("* expected diff: %#v\n", exp)
		}

		for _, got := range diffInfos {
			t.Logf("* found diff: %#v\n", got)
		}
		t.Fatalf("Found diffs do not match expected diffs!")
	}
}

func TestDiffNoDiff(t *testing.T) {
	oldDef := V2ExampleDefinition()
	newDef := V2ExampleDefinition()

	expectedDiffInfos := []DiffInfo{}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

func TestDiffServiceNameUpdated(t *testing.T) {
	oldDef := V2ExampleDefinition()
	oldDef.AppName = "service"
	newDef := V2ExampleDefinition()
	newDef.AppName = "my-new-service-name"

	expectedDiffInfos := []DiffInfo{
		DiffInfo{
			Type:   DiffTypeServiceNameUpdated,
			Action: "re-create service",
			Reason: "updating service name breaks service discovery",
			Old:    "service",
			New:    "my-new-service-name",
		},
	}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

func TestDiffComponentAdded(t *testing.T) {
	oldDef := V2ExampleDefinition()
	newDef := V2ExampleDefinition()
	newDef.Components[ComponentName("my-new-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
	}

	expectedDiffInfos := []DiffInfo{
		DiffInfo{
			Type:      DiffTypeComponentAdded,
			Component: "my-new-component",
			Action:    "add component",
			Reason:    "component 'my-new-component' not found in old definition",
			New:       "my-new-component",
		},
	}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

func TestDiffComponentRemoved(t *testing.T) {
	oldDef := V2ExampleDefinition()
	oldDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
	}
	newDef := V2ExampleDefinition()

	expectedDiffInfos := []DiffInfo{
		DiffInfo{
			Type:      DiffTypeComponentRemoved,
			Component: "my-old-component",
			Action:    "remove component",
			Reason:    "component 'my-old-component' not found in new definition",
			Old:       "my-old-component",
		},
	}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

func TestDiffComponentAddedAndRemoved(t *testing.T) {
	oldDef := V2ExampleDefinition()
	oldDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
	}
	newDef := V2ExampleDefinition()
	newDef.Components[ComponentName("my-new-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
	}

	expectedDiffInfos := []DiffInfo{
		DiffInfo{
			Type:      DiffTypeComponentAdded,
			Component: "my-new-component",
			Action:    "add component",
			Reason:    "component 'my-new-component' not found in old definition",
			New:       "my-new-component",
		},
		DiffInfo{
			Type:      DiffTypeComponentRemoved,
			Component: "my-old-component",
			Action:    "remove component",
			Reason:    "component 'my-old-component' not found in new definition",
			Old:       "my-old-component",
		},
	}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

func TestDiffComponentUpdated(t *testing.T) {
	oldDef := V2ExampleDefinition()
	oldDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
	}
	newDef := V2ExampleDefinition()
	newDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("8080/tcp")}, // port updated
	}

	expectedDiffInfos := []DiffInfo{
		DiffInfo{
			Type:      DiffTypeComponentUpdated,
			Component: "my-old-component",
			Action:    "update component",
			Reason:    "component 'my-old-component' changed in new definition",
		},
	}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

func TestDiffComponentNoImageExposeRemoved(t *testing.T) {
	oldDef := V2ExampleDefinition()
	newDef := V2ExampleDefinition()

	newComponentName := ComponentName("test-no-image")
	component := &ComponentDefinition{
		Expose: ExposeDefinitions([]ExposeDefinition{
			ExposeDefinition{
				Port:       generictypes.MustParseDockerPort("8080/tcp"),
				Component:  ComponentName("foo-bar"),
				TargetPort: generictypes.MustParseDockerPort("8080/tcp"),
			},
		}),
	}
	oldDef.Components[newComponentName] = component

	// Create a copy of the component which has a second expose
	newDefComponent := *component
	newDefComponent.Expose = append(newDefComponent.Expose, ExposeDefinition{
		Port:       generictypes.MustParseDockerPort("8081/tcp"),
		Component:  ComponentName("foo-bar2"),
		TargetPort: generictypes.MustParseDockerPort("8081/tcp"),
	})
	newDef.Components[newComponentName] = &newDefComponent

	expectedDiffInfos := []DiffInfo{
		{
			Type:      DiffTypeComponentUpdated,
			Component: "test-no-image",
			Action:    "update component",
			Reason:    "component 'test-no-image' changed in new definition",
		},
	}
	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

func TestDiffFullDefinitionUpdate(t *testing.T) {
	// The original implementation (of "diff" creation) has an issue with the go
	// implementation of map, not being consistent with respect to ordering of
	// elements while iterating through the map. With this loop we prevent that
	// it works "by mistake" the first time (but not the second or third time)
	for i := 0; i < 1000; i++ {
		rawOldDef := `
		{
			"name": "redis-example",
			"components": {
				"redis": {
					"image": "redis",
					"ports": 6379
				},
				"service": {
					"image": "giantswarm/redis-example:0.3.0",
					"ports": 80,
					"domains": { "80": "foo.com" },
					"links": [
						{ "component": "redis", "target_port": 6379 }
					]
				},
				"redis2": {
					"image": "redis",
					"ports": 6379
				},
				"service2": {
					"image": "giantswarm/redis-example:0.3.0",
					"ports": 80,
					"domains": { "80": "foo.com" },
					"links": [
						{ "component": "redis2", "target_port": 6379 }
					]
				},
				"redis3": {
					"image": "redis",
					"ports": 6379
				},
				"service3": {
					"image": "giantswarm/redis-example:0.3.0",
					"ports": 80,
					"domains": { "80": "foo.com" },
					"links": [
						{ "component": "redis3", "target_port": 6379 }
					]
				}
			}
		}
	`

		var oldDef V2AppDefinition
		if err := json.Unmarshal([]byte(rawOldDef), &oldDef); err != nil {
			t.Fatalf("failed to unmarshal service definition: %#v", err)
		}

		rawNewDef := `
		{
			"name": "redis-example-2",
			"components": {
				"redis1": {
					"image": "redis",
					"ports": 6000
				},
				"service1": {
					"image": "giantswarm/redis-example:0.3.0",
					"ports": 80,
					"domains": { "80": "bar.com" },
					"links": [
						{ "component": "redis1", "target_port": 6000 }
					]
				},
				"redis2": {
					"image": "redis",
					"ports": 6000
				},
				"service2": {
					"image": "giantswarm/redis-example:0.3.0",
					"ports": 80,
					"domains": { "80": "foo.com" },
					"links": [
						{ "component": "redis2", "target_port": 6000 }
					]
				},
				"redis3": {
					"image": "redis",
					"ports": 6379
				},
				"service3": {
					"image": "giantswarm/redis-example:0.3.0",
					"ports": 80,
					"domains": { "80": "foo.com" },
					"links": [
						{ "component": "redis3", "target_port": 6379 }
					]
				}
			}
		}
	`

		var newDef V2AppDefinition
		if err := json.Unmarshal([]byte(rawNewDef), &newDef); err != nil {
			t.Fatalf("failed to unmarshal service definition: %#v", err)
		}

		expectedDiffInfos := []DiffInfo{
			DiffInfo{
				Type:   DiffTypeServiceNameUpdated,
				Action: "re-create service",
				Reason: "updating service name breaks service discovery",
				Old:    "redis-example",
				New:    "redis-example-2",
			},
			DiffInfo{
				Type:      DiffTypeComponentAdded,
				Component: "redis1",
				Action:    "add component",
				Reason:    "component 'redis1' not found in old definition",
				New:       "redis1",
			},
			DiffInfo{
				Type:      DiffTypeComponentAdded,
				Component: "service1",
				Action:    "add component",
				Reason:    "component 'service1' not found in old definition",
				New:       "service1",
			},
			DiffInfo{
				Type:      DiffTypeComponentUpdated,
				Component: "redis2",
				Action:    "update component",
				Reason:    "component 'redis2' changed in new definition",
			},
			DiffInfo{
				Type:      DiffTypeComponentUpdated,
				Component: "service2",
				Action:    "update component",
				Reason:    "component 'service2' changed in new definition",
			},
			DiffInfo{
				Type:      DiffTypeComponentRemoved,
				Component: "redis",
				Action:    "remove component",
				Reason:    "component 'redis' not found in new definition",
				Old:       "redis",
			},
			DiffInfo{
				Type:      DiffTypeComponentRemoved,
				Component: "service",
				Action:    "remove component",
				Reason:    "component 'service' not found in new definition",
				Old:       "service",
			},
		}

		testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
	}
}

// Because of our smart data types some definitions can be represented
// differently. The following test ensures that a diff only is created in case
// the component really changed.
func TestDiffComponentDefinitionNoUpdate(t *testing.T) {
	rawOldDef := `
			{
				"name": "redis-example",
				"components": {
					"redis": {
						"image": "redis",
						"ports": [
							6379
						]
					},
					"service": {
						"image": "giantswarm/redis-example:0.3.0",
						"ports": [
							80
						],
						"domains": {
							"80/tcp": [
								"foo.com"
							]
						},
						"links": [
							{
								"component": "redis",
								"target_port": 6379
							}
						]
					}
				}
			}
	`

	var oldDef V2AppDefinition
	if err := json.Unmarshal([]byte(rawOldDef), &oldDef); err != nil {
		t.Fatalf("failed to unmarshal service definition: %#v", err)
	}

	rawNewDef := `
			{
				"name": "redis-example",
				"components": {
					"redis": {
						"image": "redis",
						"ports": 6379
					},
					"service": {
						"image": "giantswarm/redis-example:0.3.0",
						"ports": 80,
						"domains": { "80": "foo.com" },
						"links": [
							{ "component": "redis", "target_port": 6379 }
						]
					},
					"redis2": {
						"image": "redis",
						"ports": 6379
					}
				}
			}
	`

	var newDef V2AppDefinition
	if err := json.Unmarshal([]byte(rawNewDef), &newDef); err != nil {
		t.Fatalf("failed to unmarshal service definition: %#v", err)
	}

	expectedDiffInfos := []DiffInfo{
		DiffInfo{
			Type:      DiffTypeComponentAdded,
			Component: "redis2",
			Action:    "add component",
			Reason:    "component 'redis2' not found in old definition",
			New:       "redis2",
		},
	}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

// scale

func TestDiff_NoScale(t *testing.T) {
	oldDef := V2ExampleDefinition()
	oldDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
	}
	newDef := V2ExampleDefinition()
	newDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
	}

	expectedDiffInfos := []DiffInfo{}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

func TestDiff_ScaleNotChanged(t *testing.T) {
	oldDef := V2ExampleDefinition()
	oldDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 2, Max: 6},
	}
	newDef := V2ExampleDefinition()
	newDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 2, Max: 6},
	}

	expectedDiffInfos := []DiffInfo{}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

func TestDiff_ScaleChanged_MinDecreased(t *testing.T) {
	oldDef := V2ExampleDefinition()
	oldDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 2, Max: 6},
	}
	newDef := V2ExampleDefinition()
	newDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 1, Max: 6}, // scale min decreased
	}

	expectedDiffInfos := []DiffInfo{
		{
			Type:      DiffTypeComponentScaleMinDecreased,
			Component: "my-old-component",
			Action:    "store component definition",
			Reason:    "min scale of component 'my-old-component' decreased in new definition",
			Old:       "{\"min\":2,\"max\":6}",
			New:       "{\"min\":1,\"max\":6}",
		},
	}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

func TestDiff_ScaleChanged_MinIncreased(t *testing.T) {
	oldDef := V2ExampleDefinition()
	oldDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 2, Max: 6},
	}
	newDef := V2ExampleDefinition()
	newDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 3, Max: 6}, // scale min increased
	}

	expectedDiffInfos := []DiffInfo{
		{
			Type:      DiffTypeComponentScaleMinIncreased,
			Component: "my-old-component",
			Action:    "eventually scale up",
			Reason:    "scaling action will be applied depending on current instance count",
			Old:       "{\"min\":2,\"max\":6}",
			New:       "{\"min\":3,\"max\":6}",
		},
	}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

func TestDiff_ScaleChanged_MaxDecreased(t *testing.T) {
	oldDef := V2ExampleDefinition()
	oldDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 2, Max: 6},
	}
	newDef := V2ExampleDefinition()
	newDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 2, Max: 5}, // scale max decreased
	}

	expectedDiffInfos := []DiffInfo{
		{
			Type:      DiffTypeComponentScaleMaxDecreased,
			Component: "my-old-component",
			Action:    "eventually scale down",
			Reason:    "scaling action will be applied depending on current instance count",
			Old:       "{\"min\":2,\"max\":6}",
			New:       "{\"min\":2,\"max\":5}",
		},
	}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

func TestDiff_ScaleChanged_MaxIncreased(t *testing.T) {
	oldDef := V2ExampleDefinition()
	oldDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 2, Max: 6},
	}
	newDef := V2ExampleDefinition()
	newDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 2, Max: 7}, // scale min increased
	}

	expectedDiffInfos := []DiffInfo{
		{
			Type:      DiffTypeComponentScaleMaxIncreased,
			Component: "my-old-component",
			Action:    "store component definition",
			Reason:    "max scale of component 'my-old-component' increased in new definition",
			Old:       "{\"min\":2,\"max\":6}",
			New:       "{\"min\":2,\"max\":7}",
		},
	}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

func TestDiff_ScaleChanged_Full(t *testing.T) {
	oldDef := V2ExampleDefinition()
	oldDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 2, Max: 6},
	}
	newDef := V2ExampleDefinition()
	newDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 1, Max: 7, Placement: OnePerMachinePlacement}, // scale min increased
	}

	expectedDiffInfos := []DiffInfo{
		{
			Type:      DiffTypeComponentScalePlacementUpdated,
			Component: "my-old-component",
			Action:    "update component",
			Reason:    "scaling strategy of component 'my-old-component' changed in new definition",
			Old:       "{\"min\":2,\"max\":6}",
			New:       "{\"min\":1,\"max\":7,\"placement\":\"one-per-machine\"}",
		},
		{
			Type:      DiffTypeComponentScaleMinDecreased,
			Component: "my-old-component",
			Action:    "store component definition",
			Reason:    "min scale of component 'my-old-component' decreased in new definition",
			Old:       "{\"min\":2,\"max\":6}",
			New:       "{\"min\":1,\"max\":7,\"placement\":\"one-per-machine\"}",
		},
		{
			Type:      DiffTypeComponentScaleMaxIncreased,
			Component: "my-old-component",
			Action:    "store component definition",
			Reason:    "max scale of component 'my-old-component' increased in new definition",
			Old:       "{\"min\":2,\"max\":6}",
			New:       "{\"min\":1,\"max\":7,\"placement\":\"one-per-machine\"}",
		},
	}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

func TestDiff_ScaleChanged_DefaultPlacement(t *testing.T) {
	oldDef := V2ExampleDefinition()
	oldDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 2, Max: 6},
	}
	newDef := V2ExampleDefinition()
	newDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 2, Max: 6, Placement: DefaultPlacement}, // placement "changed", but stays the same because of the default
	}

	expectedDiffInfos := []DiffInfo{}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

func TestDiff_ScaleChanged_Placement(t *testing.T) {
	oldDef := V2ExampleDefinition()
	oldDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 2, Max: 6},
	}
	newDef := V2ExampleDefinition()
	newDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 2, Max: 6, Placement: OnePerMachinePlacement}, // placement "changed", but stays the same because of the default
	}

	expectedDiffInfos := []DiffInfo{
		{
			Type:      DiffTypeComponentScalePlacementUpdated,
			Component: "my-old-component",
			Action:    "update component",
			Reason:    "scaling strategy of component 'my-old-component' changed in new definition",
			Old:       "{\"min\":2,\"max\":6}",
			New:       "{\"min\":2,\"max\":6,\"placement\":\"one-per-machine\"}",
		},
	}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

func TestDiff_ScaleChanged_PortChanged(t *testing.T) {
	oldDef := V2ExampleDefinition()
	oldDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 2, Max: 6},
	}
	newDef := V2ExampleDefinition()
	newDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("88/tcp")}, // port changed
		Scale: &ScaleDefinition{Min: 2, Max: 7},                                      // scale max increased
	}

	expectedDiffInfos := []DiffInfo{
		{
			Type:      DiffTypeComponentUpdated,
			Component: "my-old-component",
			Action:    "update component",
			Reason:    "component 'my-old-component' changed in new definition",
		},
		{
			Type:      DiffTypeComponentScaleMaxIncreased,
			Component: "my-old-component",
			Action:    "store component definition",
			Reason:    "max scale of component 'my-old-component' increased in new definition",
			Old:       "{\"min\":2,\"max\":6}",
			New:       "{\"min\":2,\"max\":7}",
		},
	}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

func TestDiff_ScaleChanged_PortChanged_InOtherComponent(t *testing.T) {
	oldDef := V2ExampleDefinition()
	oldDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 2, Max: 6},
	}
	newDef := V2ExampleDefinition()
	newDef.Components[ComponentName("my-old-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
		Scale: &ScaleDefinition{Min: 3, Max: 6}, // scale min increased
	}

	// other component
	oldDef.Components[ComponentName("my-other-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
	}
	newDef.Components[ComponentName("my-other-component")] = &ComponentDefinition{
		Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("88/tcp")}, // port changed
	}

	expectedDiffInfos := []DiffInfo{
		{
			Type:      DiffTypeComponentScaleMinIncreased,
			Component: "my-old-component",
			Action:    "eventually scale up",
			Reason:    "scaling action will be applied depending on current instance count",
			Old:       "{\"min\":2,\"max\":6}",
			New:       "{\"min\":3,\"max\":6}",
		},
		{
			Type:      DiffTypeComponentUpdated,
			Component: "my-other-component",
			Action:    "update component",
			Reason:    "component 'my-other-component' changed in new definition",
		},
	}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

// TODO test each definition difference inside a component definition
