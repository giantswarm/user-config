package userconfig

import (
	"github.com/juju/errgo"

	"encoding/json"
	"regexp"
	"strings"
)

var (
	// Must have at least one .
	PatternRegistry  = regexp.MustCompile("^[A-Za-z]+[A-Za-z0-9-_]*\\.[A-Za-z]+[A-Za-z0-9-_]*$")
	PatternNamespace = regexp.MustCompile("^[a-zA-Z0-9_]{4,}$")
	PatternImage     = regexp.MustCompile("^[a-zA-Z0-9-_]{4,}$")
	PatternVersion   = regexp.MustCompile("^[a-zA-Z0-9-\\.]+$")
)

var (
	// NOTE: This format description is slightly different from what we parse down there.
	// The format given here is the one docker documents. But the repository also consist of a
	// namespace which is more or less always there. Since our business logic requires some checks based on the
	// namespace, we parse it explicitly.
	ErrInvalidFormat = errgo.New("Not a valid docker image. Format: [<registry>/]<repository>[:<version>]")
)

func ParseDockerImage(image string) (DockerImage, error) {
	var dockerImage DockerImage
	err := dockerImage.parse(image)
	return dockerImage, err
}

type DockerImage struct {
	origin string // All of the below strings combined: <registry>/<namespace>/<repository>:<version>

	Registry   string // The registry name
	Namespace  string // The namespace
	Repository string // The repository name
	Version    string // The version part
}

func (img DockerImage) MarshalJSON() ([]byte, error) {
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
	if strings.TrimSpace(input) == "" {
		return errgo.Notef(ErrInvalidFormat, "Zero length")
	}
	img.origin = input

	splitByPath := strings.Split(input, "/")
	if len(splitByPath) > 3 {
		return errgo.Notef(ErrInvalidFormat, "Too many path elements")
	}

	if isRegistry(splitByPath[0]) {
		img.Registry = splitByPath[0]
		splitByPath = splitByPath[1:]
	}

	switch len(splitByPath) {
	case 1:
		img.Repository = splitByPath[0]
	case 2:
		img.Namespace = splitByPath[0]
		img.Repository = splitByPath[1]

		if !isNamespace(img.Namespace) {
			return errgo.Notef(ErrInvalidFormat, "Invalid namespace part: "+img.Namespace)
		}
	default:
		return errgo.Notef(ErrInvalidFormat, "Invalid format")
	}

	// Now split img.Repository into img.Repository and img.Version
	splitByVersionSeparator := strings.Split(img.Repository, ":")

	switch len(splitByVersionSeparator) {
	case 2:
		img.Repository = splitByVersionSeparator[0]
		img.Version = splitByVersionSeparator[1]

		if !isVersion(img.Version) {
			return errgo.Notef(ErrInvalidFormat, "Invalid version %#v", img.Version)
		}
	case 1:
		img.Repository = splitByVersionSeparator[0]
	default:
		return errgo.Notef(ErrInvalidFormat, "Too many double colons")
	}

	if !isImage(img.Repository) {
		return errgo.Notef(ErrInvalidFormat, "Invalid image part %#v", img.Repository)
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
	return PatternRegistry.MatchString(input)
}

func isNamespace(input string) bool {
	if len(input) < 4 {
		return false
	}
	return PatternNamespace.MatchString(input)
}

func isImage(input string) bool {
	if len(input) == 0 {
		return false
	}
	return PatternImage.MatchString(input)
}

func isVersion(input string) bool {
	return PatternVersion.MatchString(input)
}
