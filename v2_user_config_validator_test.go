package userconfig_test

import (
	"encoding/json"
	"testing"

	"github.com/giantswarm/user-config"
)

func TestParseV2AppDef(t *testing.T) {
	b := []byte(`{
		"nodes": {
			"node/a": {
				"image": "registry/namespace/repository:version",
				"ports": [ "80/tcp" ],
				"links": [
					{ "name": "node/b", "port": 6379, "same_machine": true }
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

	if nodeA.Image.DockerImage().Registry != "registry" {
		t.Fatalf("invalid registry: %s", nodeA.Image.DockerImage().Registry)
	}
	if nodeA.Image.DockerImage().Namespace != "namespace" {
		t.Fatalf("invalid namespace: %s", nodeA.Image.DockerImage().Namespace)
	}
	if nodeA.Image.DockerImage().Repository != "repository" {
		t.Fatalf("invalid repository: %s", nodeA.Image.DockerImage().Repository)
	}
	if nodeA.Image.DockerImage().Version != "version" {
		t.Fatalf("invalid version: %s", nodeA.Image.DockerImage().Version)
	}

	nodeB, ok := appDef.Nodes["node/b"]
	if !ok {
		t.Fatalf("missing node")
	}

	if nodeB.Image.DockerImage().Registry != "" {
		t.Fatalf("invalid registry: %s", nodeB.Image.DockerImage().Registry)
	}
	if nodeB.Image.DockerImage().Namespace != "dockerfile" {
		t.Fatalf("invalid namespace: %s", nodeB.Image.DockerImage().Namespace)
	}
	if nodeB.Image.DockerImage().Repository != "redis" {
		t.Fatalf("invalid repository: %s", nodeB.Image.DockerImage().Repository)
	}
	if nodeB.Image.DockerImage().Version != "" {
		t.Fatalf("invalid version: %s", nodeB.Image.DockerImage().Version)
	}
}

func TestV2AppDefInvalidFieldName(t *testing.T) {
	b := []byte(`{
		"nodes": {
			"node/a": {
				"image": "registry/namespace/repository:version",
				"foo": [ "80/tcp" ]
			}
		}
	}`)

	var appDef userconfig.V2AppDefinition
	err := json.Unmarshal(b, &appDef)
	if err == nil {
		t.Fatalf("json.Unmarshal NOT failed")
	}
	if err.Error() != `Cannot parse app definition. Unknown field '["nodes"]["node/a"]["foo"]' detected.` {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsUnknownJsonField(err) {
		t.Fatalf("expetced error to be UnknownJSONFieldError")
	}
}

func TestV2AppDefFixFieldName(t *testing.T) {
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

	if nodeFooBar.Image.DockerImage().Registry != "registry" {
		t.Fatalf("invalid registry: %s", nodeFooBar.Image.DockerImage().Registry)
	}
	if nodeFooBar.Image.DockerImage().Namespace != "namespace" {
		t.Fatalf("invalid namespace: %s", nodeFooBar.Image.DockerImage().Namespace)
	}
	if nodeFooBar.Image.DockerImage().Repository != "repository" {
		t.Fatalf("invalid repository: %s", nodeFooBar.Image.DockerImage().Repository)
	}
	if nodeFooBar.Image.DockerImage().Version != "version" {
		t.Fatalf("invalid version: %s", nodeFooBar.Image.DockerImage().Version)
	}
}

func TestV2AppDefCannotFixFieldName(t *testing.T) {
	b := []byte(`{
		"nodes": {
			"node/fooBar": {
				"imaGe": "registry/namespace/repository:version"
			}
		}
	}`)

	var appDef userconfig.V2AppDefinition
	err := json.Unmarshal(b, &appDef)
	if err == nil {
		t.Fatalf("json.Unmarshal NOT failed")
	}
	if err.Error() != `Cannot parse app definition. Unknown field '["nodes"]["node/fooBar"]["ima_ge"]' detected.` {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsUnknownJsonField(err) {
		t.Fatalf("expetced error to be UnknownJSONFieldError")
	}
}
