package userconfig_test

import (
	"encoding/json"
	"testing"

	"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
)

func TestV2UnmarshalValidDomains(t *testing.T) {
	app := ExampleDefinition()
	app.Services[0].Components[0].Domains = map[generictypes.Domain]generictypes.DockerPort{
		generictypes.Domain("i.am.correct.com"):       generictypes.MustParseDockerPort("80/tcp"),
		generictypes.Domain("i.am.correct.too.com"):   generictypes.MustParseDockerPort("80/tcp"),
		generictypes.Domain("i.80.correct.too.com"):   generictypes.MustParseDockerPort("80/tcp"),
		generictypes.Domain("i.am80.correct.too.com"): generictypes.MustParseDockerPort("80/tcp"),
	}

	data, err := json.Marshal(app)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var app2 userconfig.AppDefinition
	if err := json.Unmarshal(data, &app2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
}

func TestV2UnmarshalInvalidDomains(t *testing.T) {
	app := ExampleDefinition()
	app.Services[0].Components[0].Domains = map[generictypes.Domain]generictypes.DockerPort{
		generictypes.Domain("i.am.correct.com"):  generictypes.MustParseDockerPort("80/tcp"),
		generictypes.Domain("i.$am.invalid.com"): generictypes.MustParseDockerPort("80/tcp"),
	}

	data, err := json.Marshal(app)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var app2 userconfig.AppDefinition
	if err := json.Unmarshal(data, &app2); err == nil {
		t.Fatalf("Invalid domain not detected")
	}
}

func TestV2ValidDomainValues(t *testing.T) {
	list := []struct {
		Input  string
		Result userconfig.V2DomainDefinitions
	}{
		// Original format: domain: port
		{`{ "foo.com": "8080/tcp" }`, userconfig.V2DomainDefinitions{
			generictypes.Domain("foo.com"): generictypes.MustParseDockerPort("8080"),
		}},
		{`{ "foo.com": "8081/tcp", "old.io": "8082" }`, userconfig.V2DomainDefinitions{
			generictypes.Domain("foo.com"): generictypes.MustParseDockerPort("8081"),
			generictypes.Domain("old.io"):  generictypes.MustParseDockerPort("8082"),
		}},
		// Reverse (new) format: port: domainList
		{`{ "8080": [ "foo.com" ] }`, userconfig.V2DomainDefinitions{
			generictypes.Domain("foo.com"): generictypes.MustParseDockerPort("8080"),
		}},
		{`{ "8080": "foo.com" }`, userconfig.V2DomainDefinitions{
			generictypes.Domain("foo.com"): generictypes.MustParseDockerPort("8080"),
		}},
		{`{ "8086/tcp": "foo.com" }`, userconfig.V2DomainDefinitions{
			generictypes.Domain("foo.com"): generictypes.MustParseDockerPort("8086"),
		}},
		{`{ "8080": [ "foo.com", "intel.com" ], "6800": "motorola.com" }`, userconfig.V2DomainDefinitions{
			generictypes.Domain("foo.com"):      generictypes.MustParseDockerPort("8080"),
			generictypes.Domain("intel.com"):    generictypes.MustParseDockerPort("8080"),
			generictypes.Domain("motorola.com"): generictypes.MustParseDockerPort("6800"),
		}},
	}

	for _, test := range list {
		var dds userconfig.V2DomainDefinitions
		if err := json.Unmarshal([]byte(test.Input), &dds); err != nil {
			t.Fatalf("Valid domain definitions value '%s' considered invalid because %v", test.Input, err)
		}
		if len(dds) != len(test.Result) {
			t.Fatalf("Invalid length, expected %v, got %v", len(test.Result), len(dds))
		}
		for d, p := range dds {
			expected := test.Result[d]
			if !p.Equals(expected) {
				t.Fatalf("Invalid element for domain %s, expected %v, got %v", d, expected, p)
			}
		}
	}
}

func TestV2InvalidDomainValues(t *testing.T) {
	list := []string{
		``,
		`{"field":"foo"}`,
	}

	for _, s := range list {
		var dds userconfig.V2DomainDefinitions
		if err := json.Unmarshal([]byte(s), &dds); err == nil {
			t.Fatalf("Invalid domain value '%s' considered valid", s)
		}
	}
}

func TestUnmarshalV2DomainFullService(t *testing.T) {
	// Test the validator for full services containing various
	// forms of domain definitions
	var appDef userconfig.V2AppDefinition

	byteSlice := []byte(`{
    "nodes": {
        "node1": {
        	"ports": [ "80/tcp" ],
            "image": "busybox",
            "domains": {
            	"foo.com": "80/tcp",
            	"foobar.com": "80"
            }
        }, 
        "node2": {
        	"ports": [ "80/tcp", "81/tcp" ],
            "image": "busybox",
            "domains": {
            	"80/tcp": ["mouse.com", "mickey.com"],
            	"81": "disney.com"
            }
        }
    }
}`)

	err := json.Unmarshal(byteSlice, &appDef)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
}
