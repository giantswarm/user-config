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
			Type: DiffInfoServiceNameUpdated,
			Old:  "service",
			New:  "my-new-service-name",
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
			Type: DiffInfoComponentAdded,
			New:  "my-new-component",
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
			Type: DiffInfoComponentRemoved,
			Old:  "my-old-component",
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
			Type: DiffInfoComponentAdded,
			New:  "my-new-component",
		},
		DiffInfo{
			Type: DiffInfoComponentRemoved,
			Old:  "my-old-component",
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
			Type: DiffInfoComponentUpdated,
			Old:  "my-old-component",
			New:  "my-old-component",
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
			Type: DiffInfoComponentUpdated,
			Old:  "test-no-image",
			New:  "test-no-image",
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
				Type: DiffInfoServiceNameUpdated,
				Old:  "redis-example",
				New:  "redis-example-2",
			},
			DiffInfo{
				Type: DiffInfoComponentAdded,
				New:  "redis1",
			},
			DiffInfo{
				Type: DiffInfoComponentAdded,
				New:  "service1",
			},
			DiffInfo{
				Type: DiffInfoComponentUpdated,
				Old:  "redis2",
				New:  "redis2",
			},
			DiffInfo{
				Type: DiffInfoComponentUpdated,
				Old:  "service2",
				New:  "service2",
			},
			DiffInfo{
				Type: DiffInfoComponentRemoved,
				Old:  "redis",
			},
			DiffInfo{
				Type: DiffInfoComponentRemoved,
				Old:  "service",
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
			Type: DiffInfoComponentAdded,
			New:  "redis2",
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

func TestDiff_ScaleChanged_Max(t *testing.T) {
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
		Scale: &ScaleDefinition{Min: 2, Max: 8}, // scale changed
	}

	expectedDiffInfos := []DiffInfo{
		{
			Type: DiffInfoComponentScaleMaxUpdated,
			Old:  "my-old-component",
			New:  "my-old-component",
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
		Scale: &ScaleDefinition{Min: 3, Max: 6}, // scale changed
	}

	expectedDiffInfos := []DiffInfo{
		{
			Type: DiffInfoComponentScaleUp,
			Old:  "my-old-component",
			New:  "my-old-component",
		},
	}

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
		Scale: &ScaleDefinition{Min: 1, Max: 6}, // scale changed
	}

	expectedDiffInfos := []DiffInfo{
		{
			Type: DiffInfoComponentScaleDown,
			Old:  "my-old-component",
			New:  "my-old-component",
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
			Type: DiffInfoComponentScalePlacementUpdated,
			Old:  "my-old-component",
			New:  "my-old-component",
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
		Scale: &ScaleDefinition{Min: 2, Max: 8},                                      // scale changed
	}

	expectedDiffInfos := []DiffInfo{
		{
			Type: DiffInfoComponentUpdated,
			Old:  "my-old-component",
			New:  "my-old-component",
		},
		{
			Type: DiffInfoComponentScaleMaxUpdated,
			Old:  "my-old-component",
			New:  "my-old-component",
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
		Scale: &ScaleDefinition{Min: 3, Max: 6}, // scale changed
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
			Type: DiffInfoComponentScaleUp,
			Old:  "my-old-component",
			New:  "my-old-component",
		},
		{
			Type: DiffInfoComponentUpdated,
			Old:  "my-other-component",
			New:  "my-other-component",
		},
	}

	testDiffCallWith(t, oldDef, newDef, expectedDiffInfos)
}

// TODO test each definition difference inside a component definition
