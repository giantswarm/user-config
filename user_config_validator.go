package userconfig

import (
	"strings"
)

func getExposedPort(port string) string {
	if port == "" {
		panic("Empty string as port given.")
	}
	splittedPort := strings.Split(port, "/")
	if splittedPort[0] == "" {
		panic("Expected port to be in format <number>/<protocol>. Got " + port)
	}
	return splittedPort[0]
}
