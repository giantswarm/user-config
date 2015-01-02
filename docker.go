package userconfig

import (
	"github.com/juju/errgo"

	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// https://github.com/docker/docker/blob/6d6dc2c1a1e36a2c3992186664d4513ce28e84ae/registry/registry.go#L27
	PatternNamespace = regexp.MustCompile(`^([a-z0-9_]{4,30})$`)
	PatternImage     = regexp.MustCompile(`^([a-z0-9-_.]+)$`)
	PatternVersion   = regexp.MustCompile("^[a-zA-Z0-9-\\.]+$")
)

var (
	// NOTE: This format description is slightly different from what we parse down there.
	// The format given here is the one docker documents. But the repository also consist of a
	// namespace which is more or less always there. Since our business logic requires some checks based on the
	// namespace, we parse it explicitly.
	ErrInvalidFormat = errgo.New("Not a valid docker image. Format: [<registry>/]<repository>[:<version>]")
)

func MustParseDockerImage(image string) DockerImage {
	img, err := ParseDockerImage(image)
	if err != nil {
		panic(errgo.Mask(err))
	}
	return img
}

func ParseDockerImage(image string) (DockerImage, error) {
	var dockerImage DockerImage
	err := dockerImage.parse(image)
	return dockerImage, err
}

type DockerImage struct {
	Registry   string // The registry name
	Namespace  string // The namespace
	Repository string // The repository name
	Version    string // The version part
}

func (img DockerImage) MarshalJSON() ([]byte, error) {
	return json.Marshal(img.String())
}

func (img *DockerImage) UnmarshalJSON(data []byte) error {
	var input string

	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}
	return img.parse(input)
}

func (img *DockerImage) parse(input string) error {
	if len(input) == 0 {
		return errgo.Notef(ErrInvalidFormat, "Zero length")
	}
	if strings.Contains(input, " ") {
		return errgo.Notef(ErrInvalidFormat, "No whitespaces allowed")
	}

	splitByPath := strings.Split(input, "/")
	if len(splitByPath) > 3 {
		return errgo.Notef(ErrInvalidFormat, "Too many path elements")
	}

	if containsRegistry(splitByPath) {
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
		// Don't apply latest, since we would produce an extra diff, when checking
		// for arbitrary keys in the app-config. 'latest' is not necessary anyway,
		// because the docker daemon does not pull all tags of an image, if there
		// is none given.
		img.Version = ""
	default:
		return errgo.Notef(ErrInvalidFormat, "Too many double colons")
	}

	if !isImage(img.Repository) {
		return errgo.Notef(ErrInvalidFormat, "Invalid image part %#v", img.Repository)
	}
	return nil
}

// Returns all image inforamtion combined: <registry>/<namespace>/<repository>:<version>
func (img DockerImage) String() string {
	var imageString string

	if img.Registry != "" {
		imageString += img.Registry + "/"
	}

	if img.Namespace != "" {
		imageString += img.Namespace + "/"
	}

	imageString += img.Repository

	if img.Version != "" {
		imageString += ":" + img.Version
	}

	return imageString
}

func containsRegistry(input []string) bool {
	// See https://github.com/docker/docker/blob/6d6dc2c1a1e36a2c3992186664d4513ce28e84ae/registry/registry.go#L204
	if len(input) == 1 {
		return false
	}
	if len(input) == 2 && strings.ContainsAny(input[0], ".:") {
		return true
	}
	if len(input) == 3 {
		return true
	}
	return false
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

func MustParseDockerPort(port string) DockerPort {
	var result DockerPort
	if err := parseDockerPort(port, &result); err != nil {
		panic(err.Error())
	}
	return result
}

func ParseDockerPort(port string) (DockerPort, error) {
	var result DockerPort
	if err := parseDockerPort(port, &result); err != nil {
		return result, Mask(err)
	}
	return result, nil
}

type modePortJSONFormat int

const (
	modePortJsonDocker modePortJSONFormat = 0
	modePortJsonNumber modePortJSONFormat = 1
	modePortJsonString modePortJSONFormat = 2
)

const (
	ProtocolTCP = "tcp"
	ProtocolUDP = "udp"
)

type DockerPort struct {
	// The port number.
	Port string

	// The protocol to use. "tcp" or "udp"
	Protocol string

	// How to format this port when marshalling as JSON.
	// 0 = format as string - ("port/protocol")
	// 1 = format as int - port
	// 2 = format as short port - "<port>"
	//
	// This is needed, because we need to marshal our ports the way we parsed them.
	// Otherwise the diff check in CheckForUnknownFields() would trigger when we
	// marshal `6379` as `"6379/tcp"`.
	formatJsonMode modePortJSONFormat
}

func (d DockerPort) String() string {
	return fmt.Sprintf("%s/%s", d.Port, d.Protocol)
}

func (d DockerPort) MarshalJSON() ([]byte, error) {
	switch d.formatJsonMode {
	case modePortJsonDocker:
		return json.Marshal(d.String())
	case modePortJsonNumber:
		if d.Protocol != ProtocolTCP {
			return nil, errgo.Newf("Invalid protocol for formatJsonMode=number")
		}
		i, err := strconv.Atoi(d.Port)
		if err != nil {
			return nil, Mask(err)
		}
		return json.Marshal(i)
	case modePortJsonString:
		if d.Protocol != ProtocolTCP {
			return nil, errgo.Newf("Invalid protocol for formatJsonMode=number")
		}
		return json.Marshal(d.Port)
	default:
		panic("Invalid 'formatJsonMode'")
	}
}

func (d *DockerPort) UnmarshalJSON(data []byte) error {
	wasNumber := false
	if data[0] != '"' {
		newData := []byte{}
		newData = append(newData, '"')
		newData = append(newData, data...)
		newData = append(newData, '"')

		data = newData

		wasNumber = true
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return Mask(err)
	}

	if err := parseDockerPort(s, d); err != nil {
		return Mask(err)
	}

	if wasNumber {
		d.formatJsonMode = modePortJsonNumber
	}

	return nil
}

func (d *DockerPort) Equals(other DockerPort) bool {
	return d.Port == other.Port && d.Protocol == other.Protocol
}

func parseDockerPort(input string, dp *DockerPort) error {
	s := strings.Split(input, "/")

	switch len(s) {
	case 1:
		dp.Port = s[0]
		dp.Protocol = ProtocolTCP
		dp.formatJsonMode = modePortJsonString
	case 2:
		dp.Port = s[0]
		dp.Protocol = s[1]
		dp.formatJsonMode = modePortJsonDocker
	default:
		return errgo.Newf("Invalid format, must be either <port> or <port>/<prot>, got '%s'", input)
	}

	if parsedPort, err := strconv.Atoi(dp.Port); err != nil {
		return errgo.Notef(err, "Port must be a number, got '%s'", dp.Port)
	} else if parsedPort < 1 || parsedPort > 65535 {
		return errgo.Notef(err, "Port must be a number between 1 and 65535, got '%s'", dp.Port)
	}

	switch dp.Protocol {
	case "":
		return errgo.Newf("Protocol must not be empty.")
	case ProtocolUDP:
		fallthrough
	case ProtocolTCP:
		return nil
	default:
		return errgo.Newf("Unknown protocol: '%s' in '%s'", dp.Protocol, input)
	}
}
