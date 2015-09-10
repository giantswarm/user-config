package userconfig_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/giantswarm/generic-types-go"
	. "github.com/giantswarm/user-config"
)

func ExampleDefinition() AppDefinition {
	return AppDefinition{
		AppName: "app",
		Services: []ServiceConfig{
			ServiceConfig{
				ServiceName: "service1",
				Components: []ComponentConfig{
					ComponentConfig{
						ComponentName: "service1component1",
						InstanceConfig: InstanceConfig{
							Image: generictypes.MustParseDockerImage("registry.giantswarm.io/landingpage:0.10.0"),
							Ports: []generictypes.DockerPort{generictypes.MustParseDockerPort("80/tcp")},
						},
					},
				},
			},
		},
	}
}

func testDiffCallWith(t *testing.T, oldDef, newDef V2AppDefinition, expectedDiffInfos []DiffInfo) {
	diffInfos := Diff(oldDef, newDef)

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
