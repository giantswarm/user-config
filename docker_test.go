package userconfig

import (
	"encoding/json"
	"testing"
)

func TestMarshal(t *testing.T) {
	expectedMessage := []byte(`"zeisss/static-website:1.0"`)

	image, err := ParseDockerImage("zeisss/static-website:1.0")
	if err != nil {
		t.Fatalf("Failed to parse input image: %v", err)
	}

	data, err := json.Marshal(image)
	if err != nil {
		t.Fatalf("Failed to marshal image %v: %v", image, err)
	}

	if len(expectedMessage) != len(data) {
		t.Logf("Expected: %s", string(expectedMessage))
		t.Fatalf("Different length: %v", string(data))
	}

	for i := 0; i < len(expectedMessage); i++ {
		if expectedMessage[i] != data[i] {
			t.Fatalf("serialized message differs at index %d: %v != %v", i, expectedMessage[i], data[i])
		}
	}
}

func TestWrongDockerImageParsing(t *testing.T) {
	msg := `["zeisss/static-website"]`

	var target []DockerImage

	if err := json.Unmarshal([]byte(msg), &target); err != nil {
		t.Fatalf("Json parsing failed: %v", err)
	}

	if len(target) != 1 {
		t.Fatalf("Wrong length: %d, expected 1", len(target))
	}

	if target[0].String() != "zeisss/static-website" {
		t.Fatalf("Wrong imagename: %s", target[0])
	}
}

var parsings = []struct {
	Input string

	ExpectedRegistry   string
	ExpectedNamespace  string
	ExpectedRepository string
	ExpectedVersion    string
}{
	{
		"zeisss/static-website",

		"",
		"zeisss",
		"static-website",
		"",
	},
	{
		"registry.private.giantswarm.io/sharethemeal/payment:1.0.0",

		"registry.private.giantswarm.io",
		"sharethemeal",
		"payment",
		"1.0.0",
	},

	{
		"192.168.59.103:5000/sharethemeal/payment",

		"192.168.59.103:5000",
		"sharethemeal",
		"payment",
		"",
	},
	{
		"192.168.59.103:5000/sharethemeal/payment:192.0.0",

		"192.168.59.103:5000",
		"sharethemeal",
		"payment",
		"192.0.0",
	},
	{
		"registry.private.giantswarm.io/app-service:latest",

		"registry.private.giantswarm.io",
		"",
		"app-service",
		"latest",
	},

	{
		"ruby",

		"",
		"",
		"ruby",
		"",
	},
}

func TestParsing(t *testing.T) {
	for _, data := range parsings {
		t.Logf("Input: %s", data.Input)
		image, err := ParseDockerImage(data.Input)
		if err != nil {
			t.Fatalf("Failed to parse docker image %#v: %v", data.Input, err)
		}

		if image.Registry != data.ExpectedRegistry {
			t.Fatalf("Unexpected registry: Expected '%s' but got '%s'", data.ExpectedRegistry, image.Registry)
		}
		if image.Repository != data.ExpectedRepository {
			t.Fatalf("Unexpected repository: '%s' but got '%s'", data.ExpectedRepository, image.Repository)
		}
		if image.Version != data.ExpectedVersion {
			t.Fatalf("Unexpected version: '%s' but got '%s'", data.ExpectedVersion, image.Version)
		}
	}
}

var invalidImages = []struct {
	Input string
}{
	{
		"192.168.59.103:500", // not a valid image name
	},
	{
		"abca/asd/asd", // First element is not a hostname
	},
	{
		"foo/image", // namespace too short
	},
	{
		"zeisss/static-website::latest",
	},
	{"zeisss/static-website\t"},
	{"  zeisss/static-website"},
	{"zeisss/  static-website"},
}

func TestParsingErrors(t *testing.T) {
	for _, data := range invalidImages {
		image, err := ParseDockerImage(data.Input)
		if err == nil {
			t.Fatalf("Expected error for input: %v\nBut got: %#v", data.Input, image)
		}
	}
}
