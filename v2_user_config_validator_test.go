package userconfig_test

import (
	"encoding/json"
	"testing"

	"github.com/giantswarm/user-config"
)

func TestParseV2ServiceDef(t *testing.T) {
	b := []byte(`{
		"nodes": {
			"node/a": {
				"image": "registry/namespace/repository:version",
				"ports": [ "80/tcp" ],
				"links": [
					{ "name": "node/b", "port": 6379 },
					{ "service": "otherapp", "port": 1234 }
				],
				"domains": { "test.domain.io": "80" }
			},
			"node/b": {
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

	if len(appDef.Nodes) != 2 {
		t.Fatalf("expected two nodes: %d given", len(appDef.Nodes))
	}

	nodeA, ok := appDef.Nodes["node/a"]
	if !ok {
		t.Fatalf("missing node")
	}

	if len(nodeA.Domains) != 1 {
		t.Fatalf("expected one domain: %d given", len(nodeA.Domains))
	}

	port, ok := nodeA.Domains["test.domain.io"]
	if !ok {
		t.Fatalf("missing domain")
	}
	if port.String() != "80/tcp" {
		t.Fatalf("invalid port: %s", port.String())
	}

	if nodeA.Image.Registry != "registry" {
		t.Fatalf("invalid registry: %s", nodeA.Image.Registry)
	}
	if nodeA.Image.Namespace != "namespace" {
		t.Fatalf("invalid namespace: %s", nodeA.Image.Namespace)
	}
	if nodeA.Image.Repository != "repository" {
		t.Fatalf("invalid repository: %s", nodeA.Image.Repository)
	}
	if nodeA.Image.Version != "version" {
		t.Fatalf("invalid version: %s", nodeA.Image.Version)
	}

	nodeB, ok := appDef.Nodes["node/b"]
	if !ok {
		t.Fatalf("missing node")
	}

	if nodeB.Image.Registry != "" {
		t.Fatalf("invalid registry: %s", nodeB.Image.Registry)
	}
	if nodeB.Image.Namespace != "dockerfile" {
		t.Fatalf("invalid namespace: %s", nodeB.Image.Namespace)
	}
	if nodeB.Image.Repository != "redis" {
		t.Fatalf("invalid repository: %s", nodeB.Image.Repository)
	}
	if nodeB.Image.Version != "" {
		t.Fatalf("invalid version: %s", nodeB.Image.Version)
	}
}

func TestV2ServiceDefFixFieldName(t *testing.T) {
	b := []byte(`{
		"Nodes": {
			"node/fooBar": {
				"Image": "registry/namespace/repository:version"
			}
		}
	}`)

	var appDef userconfig.V2AppDefinition
	err := json.Unmarshal(b, &appDef)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	nodeFooBar, ok := appDef.Nodes["node/fooBar"]
	if !ok {
		t.Fatalf("missing node")
	}

	if nodeFooBar.Image.Registry != "registry" {
		t.Fatalf("invalid registry: %s", nodeFooBar.Image.Registry)
	}
	if nodeFooBar.Image.Namespace != "namespace" {
		t.Fatalf("invalid namespace: %s", nodeFooBar.Image.Namespace)
	}
	if nodeFooBar.Image.Repository != "repository" {
		t.Fatalf("invalid repository: %s", nodeFooBar.Image.Repository)
	}
	if nodeFooBar.Image.Version != "version" {
		t.Fatalf("invalid version: %s", nodeFooBar.Image.Version)
	}
}

func TestV2ServiceDefCannotFixFieldName(t *testing.T) {
	b := []byte(`{
		"nodes": {
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
	if err.Error() != `unknown JSON field: ["nodes"]["foo/bar"]["ima_ge"]` {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsUnknownJsonField(err) {
		t.Fatalf("expetced error to be UnknownJSONFieldError")
	}
}

func TestUnmarshalV2ServiceDefMissingField(t *testing.T) {
	// "image" is missing
	b := []byte(`{
		"nodes": {
			"foo/bar": {}
		}
	}`)

	var appDef userconfig.V2AppDefinition
	err := json.Unmarshal(b, &appDef)
	if err == nil {
		t.Fatalf("json.Unmarshal NOT failed")
	}
	if err.Error() != `node 'foo/bar' must have an 'image'` {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsInvalidNodeDefinition(err) {
		t.Fatalf("expected error to be MissingJSONFieldError")
	}
}

func TestUnmarshalV2ServiceDefUnknownField(t *testing.T) {
	b := []byte(`{
		"nodes": {
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
	if err.Error() != `unknown JSON field: ["nodes"]["foo/bar"]["unknown"]` {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsUnknownJsonField(err) {
		t.Fatalf("expetced error to be UnknownJSONFieldError")
	}
}
