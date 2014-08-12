package userconfig

import (
	"strings"
)

// input: "serviceName", "session",                 output: "serviceName"     "session"
// input: "serviceName", "session-service/session", output: "session-service" "session"
func ParseDependency(serviceName, dependency string) (string, string) {
	slashSplitted := strings.Split(dependency, "/")

	depServiceName := ""
	depComponentName := ""

	if len(slashSplitted) == 1 {
		depServiceName = serviceName
		depComponentName = dependency
	} else {
		depServiceName = slashSplitted[0]
		depComponentName = slashSplitted[1]
	}

	return depServiceName, depComponentName
}
