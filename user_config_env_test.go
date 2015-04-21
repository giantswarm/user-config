package userconfig_test

import (
	"encoding/json"
	"fmt"
	"testing"

	userConfigPkg "github.com/giantswarm/user-config"
)

func TestUnmarshalEnvArray(t *testing.T) {
	var compConfig userConfigPkg.ComponentConfig

	byteSlice := []byte(`{
        "component_name": "api",
        "image": "registry/namespace/repository:version",
        "env": [ "key1=value1", "key2=value2" ],
        "domains": { "test.domain.io": "80" }
		        }`)

	err := json.Unmarshal(byteSlice, &compConfig)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := fmt.Sprintf("%v", compConfig.Env)
	expected := "[key1=value1 key2=value2]"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

func TestUnmarshalEnvStruct(t *testing.T) {
	for i := 0; i < 10000; i++ {
		var compConfig userConfigPkg.ComponentConfig

		byteSlice := []byte(`{
        "component_name": "api",
        "image": "registry/namespace/repository:version",
        "env": { "key1": "value1", "key2": "value2" },
        "domains": { "test.domain.io": "80" }
		        }`)

		err := json.Unmarshal(byteSlice, &compConfig)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		got := fmt.Sprintf("%v", compConfig.Env)
		expected := "[key1=value1 key2=value2]"
		if got != expected {
			t.Fatalf("Invalid result: got %s, expected %s", got, expected)
		}
	}
}

func TestUnmarshalEnvArrayEmpty(t *testing.T) {
	var compConfig userConfigPkg.ComponentConfig

	byteSlice := []byte(`{
        "component_name": "api",
        "image": "registry/namespace/repository:version",
        "env": [],
        "domains": { "test.domain.io": "80" }
		        }`)

	err := json.Unmarshal(byteSlice, &compConfig)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := fmt.Sprintf("%v", compConfig.Env)
	expected := "[]"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

func TestUnmarshalEnvStructEmpty(t *testing.T) {
	var compConfig userConfigPkg.ComponentConfig

	byteSlice := []byte(`{
        "component_name": "api",
        "image": "registry/namespace/repository:version",
        "env": {},
        "domains": { "test.domain.io": "80" }
		        }`)

	err := json.Unmarshal(byteSlice, &compConfig)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := fmt.Sprintf("%v", compConfig.Env)
	expected := "[]"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

func TestUnmarshalEnvFullApp(t *testing.T) {
	// Test the validator for full apps containing both array and structs
	var appConfig userConfigPkg.AppConfig

	byteSlice := []byte(`{
    "app_name": "envtest",
    "services": [{
        "service_name": "envtest-service",
        "components": [{
            "component_name": "env-array",
            "env": [
                "KEY=env-array"
            ],
            "image": "busybox",
            "args": ["sh", "-c", "while true; do echo \"Beep $KEY\"; sleep 2; done"]
        }, {
            "component_name": "env-struct",
            "env": {
                "KEY": "env-struct"
            },
            "image": "busybox",
            "args": ["sh", "-c", "while true; do echo \"Beep $KEY\"; sleep 2; done"]
        }]
    }]
}`)

	err := json.Unmarshal(byteSlice, &appConfig)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
}

func TestUnmarshalEnvFullAppUpperCase(t *testing.T) {
	// Test the validator for full apps containing both array and structs with uppercase env keys
	var appConfig userConfigPkg.AppDefinition

	byteSlice := []byte(`{
    "app_name": "envtest",
    "services": [{
        "service_name": "envtest-service",
        "components": [{
            "component_name": "env-struct",
            "image": "busybox",
            "env": [
                "KEY=env-array"
            ]
        }, {
            "component_name": "env-struct",
            "image": "busybox",
            "env": {
                "KEY": "env-struct"
            }
        }]
    }]
}`)

	err := json.Unmarshal(byteSlice, &appConfig)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := fmt.Sprintf("%v", appConfig)
	expected := "{envtest map[] [{envtest-service map[] <nil> [{env-struct <nil> {busybox  [] [KEY=env-array] [] [] map[] []}} {env-struct <nil> {busybox  [] [KEY=env-struct] [] [] map[] []}}]}]}"
	if got != expected {
		t.Fatalf("Invalid result: got \n%s\nexpected\n%s", got, expected)
	}
}
