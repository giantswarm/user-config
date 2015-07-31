package userconfig_test

import (
	"encoding/json"
	"testing"

	"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
)

func TestUnmarshalValidDomains(t *testing.T) {
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

func TestUnmarshalInvalidDomains(t *testing.T) {
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