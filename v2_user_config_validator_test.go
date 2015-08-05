package userconfig_test

import (
	"encoding/json"
	"testing"

	"github.com/giantswarm/user-config"
)

func TestParseV2AppDef(t *testing.T) {
	b := []byte(`{
		"components": {
			"component/a": {
				"image": "registry/namespace/repository:version",
				"ports": [ "80/tcp" ],
				"links": [
					{ "component": "component/b", "target_port": 6379 },
					{ "service": "otherapp", "target_port": 1234 }
				],
				"domains": { "test.domain.io": "80" }
			},
			"component/b": {
				"image": "dockerfile/redis",
				"ports": [ "6379/tcp" ],
				"volumes": [
					{ "path": "/data", "size": "5 GB" },
					{ "path": "/data2", "size": "8" },
					{ "path": "/data3", "size": "8  G" },
					{ "path": "/data4", "size": "8GB" }
				]
			}
		}
	}`)

	var appDef userconfig.V2AppDefinition
	err := json.Unmarshal(b, &appDef)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if len(appDef.Components) != 2 {
		t.Fatalf("expected two components: %d given", len(appDef.Components))
	}

	componentA, ok := appDef.Components["component/a"]
	if !ok {
		t.Fatalf("missing component")
	}

	if len(componentA.Domains) != 1 {
		t.Fatalf("expected one domain: %d given", len(componentA.Domains))
	}

	port, ok := componentA.Domains["test.domain.io"]
	if !ok {
		t.Fatalf("missing domain")
	}
	if port[0].String() != "80/tcp" {
		t.Fatalf("invalid port: %s", port[0].String())
	}

	if componentA.Image.Registry != "registry" {
		t.Fatalf("invalid registry: %s", componentA.Image.Registry)
	}
	if componentA.Image.Namespace != "namespace" {
		t.Fatalf("invalid namespace: %s", componentA.Image.Namespace)
	}
	if componentA.Image.Repository != "repository" {
		t.Fatalf("invalid repository: %s", componentA.Image.Repository)
	}
	if componentA.Image.Version != "version" {
		t.Fatalf("invalid version: %s", componentA.Image.Version)
	}

	componentB, ok := appDef.Components["component/b"]
	if !ok {
		t.Fatalf("missing component")
	}

	if componentB.Image.Registry != "" {
		t.Fatalf("invalid registry: %s", componentB.Image.Registry)
	}
	if componentB.Image.Namespace != "dockerfile" {
		t.Fatalf("invalid namespace: %s", componentB.Image.Namespace)
	}
	if componentB.Image.Repository != "redis" {
		t.Fatalf("invalid repository: %s", componentB.Image.Repository)
	}
	if componentB.Image.Version != "" {
		t.Fatalf("invalid version: %s", componentB.Image.Version)
	}
}

func TestV2AppDefFixFieldName(t *testing.T) {
	b := []byte(`{
		"Components": {
			"component/fooBar": {
				"Image": "registry/namespace/repository:version"
			}
		}
	}`)

	var appDef userconfig.V2AppDefinition
	err := json.Unmarshal(b, &appDef)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	componentFooBar, ok := appDef.Components["component/fooBar"]
	if !ok {
		t.Fatalf("missing component")
	}

	if componentFooBar.Image.Registry != "registry" {
		t.Fatalf("invalid registry: %s", componentFooBar.Image.Registry)
	}
	if componentFooBar.Image.Namespace != "namespace" {
		t.Fatalf("invalid namespace: %s", componentFooBar.Image.Namespace)
	}
	if componentFooBar.Image.Repository != "repository" {
		t.Fatalf("invalid repository: %s", componentFooBar.Image.Repository)
	}
	if componentFooBar.Image.Version != "version" {
		t.Fatalf("invalid version: %s", componentFooBar.Image.Version)
	}
}

func TestV2AppDefCannotFixFieldName(t *testing.T) {
	b := []byte(`{
		"components": {
			"foo/bar": {
				"imaGe": "registry/namespace/repository:version"
			}
		}
	}`)

	var appDef userconfig.V2AppDefinition
	err := json.Unmarshal(b, &appDef)
	if err == nil {
		t.Fatalf("json.Unmarshal NOT failed")
	}
	if err.Error() != `unknown JSON field: ["components"]["foo/bar"]["ima_ge"]` {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsUnknownJsonField(err) {
		t.Fatalf("expetced error to be UnknownJSONFieldError")
	}
}

func TestUnmarshalV2AppDefMissingField(t *testing.T) {
	// "image" is missing
	b := []byte(`{
		"components": {
			"foo/bar": {}
		}
	}`)

	var appDef userconfig.V2AppDefinition
	err := json.Unmarshal(b, &appDef)
	if err == nil {
		t.Fatalf("json.Unmarshal NOT failed")
	}
	if err.Error() != `component 'foo/bar' must have an 'image'` {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsInvalidComponentDefinition(err) {
		t.Fatalf("expected error to be MissingJSONFieldError")
	}
}

func TestUnmarshalV2AppDefUnknownField(t *testing.T) {
	b := []byte(`{
		"components": {
			"foo/bar": {
				"image": "registry/namespace/repository:version",
				"unknown": "unknown"
			}
		}
	}`)

	var appDef userconfig.V2AppDefinition
	err := json.Unmarshal(b, &appDef)
	if err == nil {
		t.Fatalf("json.Unmarshal NOT failed")
	}
	if err.Error() != `unknown JSON field: ["components"]["foo/bar"]["unknown"]` {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsUnknownJsonField(err) {
		t.Fatalf("expetced error to be UnknownJSONFieldError")
	}
}
