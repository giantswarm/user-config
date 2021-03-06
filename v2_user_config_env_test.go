package userconfig_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/giantswarm/user-config"
)

func TestUnmarshalV2EnvArray(t *testing.T) {
	var componentDef userconfig.ComponentDefinition

	byteSlice := []byte(`{ "env": [ "key1=value1", "key2=value2" ] }`)

	err := json.Unmarshal(byteSlice, &componentDef)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := fmt.Sprintf("%v", componentDef.Env)
	expected := "[key1=value1 key2=value2]"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

func TestUnmarshalV2EnvStruct(t *testing.T) {
	// The original implementation (of "env" parsing) has an issue with the go
	// implementation of map, not being consistent with respect to ordering of
	// elements. With this loop we prevent that it works "by mistake" the first
	// time (but not the second or third time)
	for i := 0; i < 1000; i++ {
		var componentDef userconfig.ComponentDefinition

		byteSlice := []byte(`{ "env": { "key1": "value1", "key2": "value2" } }`)

		err := json.Unmarshal(byteSlice, &componentDef)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		got := fmt.Sprintf("%v", componentDef.Env)
		expected := "[key1=value1 key2=value2]"
		if got != expected {
			t.Fatalf("Invalid result: got %s, expected %s", got, expected)
		}
	}
}

func TestUnmarshalV2EnvArrayEmpty(t *testing.T) {
	var componentDef userconfig.ComponentDefinition

	byteSlice := []byte(`{ "env": [] }`)

	err := json.Unmarshal(byteSlice, &componentDef)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := fmt.Sprintf("%v", componentDef.Env)
	expected := "[]"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

func TestUnmarshalV2EnvStructEmpty(t *testing.T) {
	var componentDef userconfig.ComponentDefinition

	byteSlice := []byte(`{ "env": {} }`)

	err := json.Unmarshal(byteSlice, &componentDef)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	got := fmt.Sprintf("%v", componentDef.Env)
	expected := "[]"
	if got != expected {
		t.Fatalf("Invalid result: got %s, expected %s", got, expected)
	}
}

func TestUnmarshalV2EnvFullService(t *testing.T) {
	// Test the validator for full services containing both array and structs
	var serviceDef userconfig.ServiceDefinition

	byteSlice := []byte(`{
    "components": {
        "env-array": {
            "env": [
                "KEY=env-array"
            ],
            "image": "busybox",
            "args": ["sh", "-c", "while true; do echo \"Beep $KEY\"; sleep 2; done"]
        },
        "env-struct": {
            "env": {
                "KEY": "env-struct"
            },
            "image": "busybox",
            "args": ["sh", "-c", "while true; do echo \"Beep $KEY\"; sleep 2; done"]
        }
    }
}`)

	err := json.Unmarshal(byteSlice, &serviceDef)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
}

func TestUnmarshalV2EnvFullServiceUpperCase(t *testing.T) {
	// Test the validator for full services containing both array and structs with uppercase env keys
	var serviceDef userconfig.ServiceDefinition

	byteSlice := []byte(`{
    "components": {
        "env-array": {
            "image": "busybox",
            "env": [
                "KEY=env-array",
                "k=v"
            ]
        },
        "env-struct": {
            "image": "busybox",
            "env": {
                "KEY": "env-struct",
                "other": "value"
            }
        }
    }
}`)

	err := json.Unmarshal(byteSlice, &serviceDef)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	envArray, err := serviceDef.Components.ComponentByName("env-array")
	if err != nil {
		t.Fatalf("ComponentByName failed: %v", err)
	}

	got := strings.Join(envArray.Env, ", ")
	expected := "KEY=env-array, k=v"
	if got != expected {
		t.Fatalf("Invalid result: got \n%s\nexpected\n%s", got, expected)
	}

	envStruct, err := serviceDef.Components.ComponentByName("env-struct")
	if err != nil {
		t.Fatalf("ComponentByName failed: %v", err)
	}

	got = strings.Join(envStruct.Env, ", ")
	expected = "KEY=env-struct, other=value"
	if got != expected {
		t.Fatalf("Invalid result: got \n%s\nexpected\n%s", got, expected)
	}
}
