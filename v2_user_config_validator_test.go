package userconfig_test

import (
	"encoding/json"
	"reflect"
	"testing"

	//"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
)

func TestMarshalUnmarshalV2AppDef(t *testing.T) {
	a := V2ExampleDefinition()

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.V2AppDefinition
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if !reflect.DeepEqual(a, b) {
		t.Fatalf("objects differ:\n%v\n%v", a, b)
	}
}

func TestParseV2AppDef(t *testing.T) {
	b := []byte(`{
		"nodes": {
			"node/a": {
				"image": "registry/namespace/repository:version",
				"ports": [ "80/tcp" ],
				"links": [
					{ "name": "redis", "port": 6379, "same_machine": true }
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

func TestV2AppDefInvalidVolumeSizeUnit(t *testing.T) {
	a := V2ExampleDefinitionWithVolume([]string{"/data"}, []string{"5 KB"})

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.V2AppDefinition
	err = json.Unmarshal(raw, &b)
	if err == nil {
		t.Fatalf("json.Unmarshal NOT failed")
	}

	if err.Error() != "Cannot parse app config. Invalid size '5 KB' detected." {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
}

func TestV2AppDefInvalidVolumeNegativeSize(t *testing.T) {
	a := V2ExampleDefinitionWithVolume([]string{"/data"}, []string{"-5 GB"})

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.V2AppDefinition
	err = json.Unmarshal(raw, &b)
	if err == nil {
		t.Fatalf("json.Unmarshal NOT failed")
	}

	if err.Error() != "Cannot parse app config. Invalid size '-5 GB' detected." {
		t.Fatalf("expected proper error, got: %s", err.Error())
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
	if !userconfig.IsErrUnknownJsonField(err) {
		t.Fatalf("expetced error to be ErrUnknownJSONField")
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
	if !userconfig.IsErrUnknownJsonField(err) {
		t.Fatalf("expetced error to be ErrUnknownJSONField")
	}
}

func TestV2AppDefInvalidVolumeDuplicatedPath(t *testing.T) {
	a := V2ExampleDefinitionWithVolume([]string{"/data", "/data"}, []string{"5 GB", "10 GB"})

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.V2AppDefinition
	err = json.Unmarshal(raw, &b)
	if err == nil {
		t.Fatalf("json.Unmarshal NOT failed")
	}

	if err.Error() != "Cannot parse app definition. Duplicate volume '/data' detected." {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsErrDuplicateVolumePath(err) {
		t.Fatalf("expetced error to be ErrDuplicateVolumePath")
	}
}
