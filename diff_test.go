package userconfig_test

import (
	"reflect"
	"testing"

	. "github.com/giantswarm/user-config"
)

func ExampleConfig() AppConfig {
	return AppConfig{
		AppName: "app",
		Services: []ServiceConfig{
			ServiceConfig{
				ServiceName: "service1",
				Components: []ComponentConfig{
					ComponentConfig{
						ComponentName: "service1component1",
						InstanceConfig: InstanceConfig{
							Image: "registry.giantswarm.io/landingpage:0.10.0",
							Ports: []string{"80/tcp"},
						},
					},
				},
			},
		},
	}
}

func testDiffCallWith(t *testing.T, newCfg, oldCfg AppConfig, expectedInfos []DiffInfo) {
	infos := Diff(newCfg, oldCfg)

	if len(infos) != len(expectedInfos) {
		t.Fatalf("Expected %d item, got %d: %v", len(expectedInfos), len(infos), infos)
	}

	if !reflect.DeepEqual(infos, expectedInfos) {
		for _, exp := range expectedInfos {
			t.Logf("* expected diff: %v\n", exp)
		}

		for _, got := range infos {
			t.Logf("* found diff: %v\n", got)
		}
		t.Fatalf("Found diffs do not match expected diffs!")
	}
}

func TestDiffAppRename(t *testing.T) {
	oldCfg := ExampleConfig()
	newCfg := ExampleConfig()
	newCfg.AppName = "app#changed"

	expectedDiffItems := []DiffInfo{
		DiffInfo{Type: InfoAppNameChanged, Name: []string{"app"}},
	}

	testDiffCallWith(t, newCfg, oldCfg, expectedDiffItems)
}

func TestDiffServiceRename(t *testing.T) {
	oldCfg := ExampleConfig()
	newCfg := ExampleConfig()
	newCfg.Services[0].ServiceName = "service1#changed"

	expectedDiffItems := []DiffInfo{
		DiffInfo{Type: InfoNodeAdded, Name: []string{"app", "service1#changed"}},
		DiffInfo{Type: InfoNodeRemoved, Name: []string{"app", "service1"}},
	}

	testDiffCallWith(t, newCfg, oldCfg, expectedDiffItems)
}

func TestDiffComponentRename(t *testing.T) {
	oldCfg := ExampleConfig()
	newCfg := ExampleConfig()
	newCfg.Services[0].Components[0].ComponentName = "service1component1#changed"

	expectedDiffItems := []DiffInfo{
		DiffInfo{Type: InfoNodeAdded, Name: []string{"app", "service1", "service1component1#changed"}},
		DiffInfo{Type: InfoNodeRemoved, Name: []string{"app", "service1", "service1component1"}},
	}

	testDiffCallWith(t, newCfg, oldCfg, expectedDiffItems)
}

func TestDiffComponentUpdateImage(t *testing.T) {
	oldCfg := ExampleConfig()
	newCfg := ExampleConfig()
	newCfg.Services[0].Components[0].Image = "landingpage2"

	expectedDiffItems := []DiffInfo{
		DiffInfo{Type: InfoInstanceConfigUpdated, Name: []string{"app", "service1", "service1component1"}},
	}

	testDiffCallWith(t, newCfg, oldCfg, expectedDiffItems)
}
func TestDiffComponentUpdateArgs(t *testing.T) {
	oldCfg := ExampleConfig()
	newCfg := ExampleConfig()
	newCfg.Services[0].Components[0].Args = []string{"--env=test"}

	expectedDiffItems := []DiffInfo{
		DiffInfo{Type: InfoInstanceConfigUpdated, Name: []string{"app", "service1", "service1component1"}},
	}

	testDiffCallWith(t, newCfg, oldCfg, expectedDiffItems)
}

func TestDiffComponentScalingChanged(t *testing.T) {
	oldCfg := ExampleConfig()
	newCfg := ExampleConfig()
	newCfg.Services[0].Components[0].ScalingPolicy = &ScalingPolicyConfig{Min: 0, Max: 1000}

	expectedDiffItems := []DiffInfo{
		DiffInfo{Type: InfoComponentScalingUpdated, Name: []string{"app", "service1", "service1component1"}},
	}

	testDiffCallWith(t, newCfg, oldCfg, expectedDiffItems)
}

// Changes in the service should not be alert if the e.g. the service was renamed (yeah, illogical, I know)
func TestDiffLowerLevelChangesShouldBeIgnored(t *testing.T) {
	oldCfg := ExampleConfig()
	newCfg := ExampleConfig()

	newCfg.Services[0].Components = append(newCfg.Services[0].Components, ComponentConfig{
		ComponentName: "service1component1",
		InstanceConfig: InstanceConfig{
			Image: "foobar",
		},
	})
	newCfg.Services[0].ServiceName = "service1#changed"

	expectedDiffItems := []DiffInfo{
		DiffInfo{Type: InfoNodeAdded, Name: []string{"app", "service1#changed"}},
		DiffInfo{Type: InfoNodeRemoved, Name: []string{"app", "service1"}},
	}

	testDiffCallWith(t, newCfg, oldCfg, expectedDiffItems)
}

func TestDiffMultipleChanges(t *testing.T) {
	oldCfg := ExampleConfig()
	newCfg := ExampleConfig()

	newCfg.Services[0].Components[0].ComponentName = "service1component1#changed"
	newCfg.Services[0].Components = append(newCfg.Services[0].Components, ComponentConfig{
		ComponentName: "service1component2",
		InstanceConfig: InstanceConfig{
			Image: "foobar",
		},
	})

	expectedDiffItems := []DiffInfo{
		DiffInfo{Type: InfoNodeAdded, Name: []string{"app", "service1", "service1component1#changed"}},
		DiffInfo{Type: InfoNodeAdded, Name: []string{"app", "service1", "service1component2"}},
		DiffInfo{Type: InfoNodeRemoved, Name: []string{"app", "service1", "service1component1"}},
	}

	testDiffCallWith(t, newCfg, oldCfg, expectedDiffItems)
}

func TestDiffComponentMultipleChanges(t *testing.T) {
	oldCfg := ExampleConfig()
	newCfg := ExampleConfig()

	newCfg.Services[0].Components[0].ScalingPolicy = &ScalingPolicyConfig{Min: 10}
	newCfg.Services[0].Components[0].InstanceConfig.Image = "new-site"

	expectedDiffItems := []DiffInfo{
		DiffInfo{Type: InfoInstanceConfigUpdated, Name: []string{"app", "service1", "service1component1"}},
		DiffInfo{Type: InfoComponentScalingUpdated, Name: []string{"app", "service1", "service1component1"}},
	}

	testDiffCallWith(t, newCfg, oldCfg, expectedDiffItems)
}
