package userconfig

import (
	"github.com/juju/errgo"

	"encoding/json"
	"regexp"
	"strings"
)

var (
	PatternNamespace = regexp.MustCompile("[a-z0-9_]")
)

var (
	ErrInvalidFormat = errgo.New("Not a valid docker image")
)

func NewDockerImage(image string) (DockerImage, error) {
	var dockerImage DockerImage
	err := dockerImage.parse(image)
	return dockerImage, err
}

type DockerImage struct {
	origin string // All of the below strings combined: <registry>/<repository>:<version>

	Registry   string // The registry name
	Repository string // The namespace and repository name
	Version    string // The version part
}

func (img *DockerImage) MarshalJSON() ([]byte, error) {
	return json.Marshal(img.origin)
}

func (img *DockerImage) UnmarshalJSON(data []byte) error {
	var input string

	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}
	return img.parse(input)
}

func (img *DockerImage) parse(input string) error {
	if strings.Trimspace(input) == "" {
		return errgo.Notef(ErrInvalidFormat, "Zero length")
	}
	img.origin = input

	splitByPath := strings.Split(input, "/")
	if len(splitByPath) > 3 {
		return errgo.Notef(ErrInvalidFormat, "Too many path elements")
	}

	index := 0
	if isRegistry(splitByPath[index]) {
		img.Registry = splitByPath[index]
		index++
	}

	if isNamespace(splitByPath[index]) {

	}

	switch len(splitByPath) {
	case 0:
		return errgo.Notef(ErrInvalidFormat, "No path separator found.")
	case 1:
		img.Repository = input
		if isRegistry(img.Repository) {
			return errgo.Notef(ErrInvalidFormat, "Only element is a registry")
		}
	case 2:
		if isRegistry(splitByPath[0]) {
			img.Registry = splitByPath[0]
			img.Repository = splitByPath[1]
		} else {
			img.Repository = input
		}
	}

	// Now split img.Image into img.Image and img.Version
	splitByVersionSeparator := strings.Split(img.Repository, ":")
	if len(splitByVersionSeparator) > 2 {
		return errgo.Notef(ErrInvalidFormat, "Too many double colons")
	}

	if len(splitByVersionSeparator) == 2 {
		img.Repository = splitByVersionSeparator[0]
		img.Version = splitByVersionSeparator[1]
	}

	if strings.TrimSpace(img.Repository) == "" {
		return errgo.Notef(ErrInvalidFormat, "No repository provided.")
	}
	return nil
}

func (img DockerImage) String() string {
	return img.origin
}

func isRegistry(input string) bool {
	if strings.ContainsAny(input, ".:") {
		return true
	}
	return false
}

// Only [a-z0-9_]
func isNamespace(input string) bool {
	if len(input) < 4 {
		return false
	}
	return PatternNamespace.MatchString(input)
}
