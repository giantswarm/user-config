package userconfig_test

import (
	"encoding/json"
	"testing"

	"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
)

func TestV2ValidDomainValues(t *testing.T) {
	list := []struct {
		Input  string
		Result userconfig.V2DomainDefinitions
	}{
		// Original format: domain: port
		{`{ "foo.com": "8080/tcp" }`, userconfig.V2DomainDefinitions{
			generictypes.Domain("foo.com"): userconfig.PortDefinitions{generictypes.MustParseDockerPort("8080")},
		}},
		{`{ "foo.com": "8081/tcp", "old.io": "8082" }`, userconfig.V2DomainDefinitions{
			generictypes.Domain("foo.com"): userconfig.PortDefinitions{generictypes.MustParseDockerPort("8081")},
			generictypes.Domain("old.io"):  userconfig.PortDefinitions{generictypes.MustParseDockerPort("8082")},
		}},
		// Reverse (new) format: port: domainList
		{`{ "8080": [ "foo.com" ] }`, userconfig.V2DomainDefinitions{
			generictypes.Domain("foo.com"): userconfig.PortDefinitions{generictypes.MustParseDockerPort("8080")},
		}},
		{`{ "8080": "foo.com" }`, userconfig.V2DomainDefinitions{
			generictypes.Domain("foo.com"): userconfig.PortDefinitions{generictypes.MustParseDockerPort("8080")},
		}},
		{`{ "8086/tcp": "foo.com" }`, userconfig.V2DomainDefinitions{
			generictypes.Domain("foo.com"): userconfig.PortDefinitions{generictypes.MustParseDockerPort("8086")},
		}},
		{`{ "8080": [ "foo.com", "intel.com" ], "6800": "motorola.com" }`, userconfig.V2DomainDefinitions{
			generictypes.Domain("foo.com"):      userconfig.PortDefinitions{generictypes.MustParseDockerPort("8080")},
			generictypes.Domain("intel.com"):    userconfig.PortDefinitions{generictypes.MustParseDockerPort("8080")},
			generictypes.Domain("motorola.com"): userconfig.PortDefinitions{generictypes.MustParseDockerPort("6800")},
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
		for d, ports := range dds {
			for _, p := range ports {
				expected := test.Result[d][0]
				if !p.Equals(expected) {
					t.Fatalf("Invalid element for domain %s, expected %v, got %v", d, expected, p)
				}
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
	//
	// The original implementation (of "domain" parsing) has an issue with the go
	// implementation of map, not being consistent with respect to ordering of
	// elements.  With this loop we prevent that it works "by mistake" the first
	// time (but not the second or third time)
	for i := 0; i < 1000; i++ {
		var appDef userconfig.V2AppDefinition

		byteSlice := []byte(`
			{
				"components": {
					"component1": {
						"domains": {
							"ape.org": "80/tcp",
							"foo.com": "80/tcp",
							"foobar.com": "80",
							"int.com": 80
						},
						"image": "busybox",
						"ports": [
							"80/tcp"
						]
					},
					"component2": {
						"domains": {
							"80/tcp": [
								"mouse.com",
								"ape.org",
								"mickey.com"
							],
							"81": "disney.com"
						},
						"image": "busybox",
						"ports": [
							"80/tcp",
							"81/tcp"
						]
					}
				}
			}
		`)

		err := json.Unmarshal(byteSlice, &appDef)
		if err != nil {
			t.Fatalf("Unmarshal failed: %#v", err)
		}
	}
}

// TestDomainStringReliability ensures that V2DomainDefinitions.String always returns
// the same string. We use this to create proper diffs between two definitions,
// so this is quiet critical.
func TestDomainStringReliability(t *testing.T) {
	dds := userconfig.V2DomainDefinitions{
		generictypes.Domain("foo.com"): userconfig.PortDefinitions{
			generictypes.MustParseDockerPort("8080"),
			generictypes.MustParseDockerPort("111"),
			generictypes.MustParseDockerPort("9999"),
		},
		generictypes.Domain("aaa.com"): userconfig.PortDefinitions{
			generictypes.MustParseDockerPort("111"),
			generictypes.MustParseDockerPort("9999"),
			generictypes.MustParseDockerPort("55"),
		},
		generictypes.Domain("zabe.com"): userconfig.PortDefinitions{
			generictypes.MustParseDockerPort("3333"),
			generictypes.MustParseDockerPort("9283"),
			generictypes.MustParseDockerPort("55"),
		},
	}

	expected := dds.String()

	for i := 0; i < 1000; i++ {
		generated := dds.String()

		if expected != generated {
			t.Log("expected domain definitions to be qual")
			t.Logf("epxected: %s", expected)
			t.Fatalf("got: %s", generated)
		}
	}
}
