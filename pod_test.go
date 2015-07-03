package userconfig_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/giantswarm/user-config"
)

func TestValidPodValues(t *testing.T) {
	list := []string{
		"none",
		"children",
		"inherit",
	}

	for _, s := range list {
		var pe struct {
			Pod userconfig.PodEnum
		}
		data := fmt.Sprintf(`{"pod": "%s"}`, s)
		if err := json.Unmarshal([]byte(data), &pe); err != nil {
			t.Fatalf("Valid pod value '%s' considered invalid because %v", s, err)
		}
	}
}

func TestInvalidPodValues(t *testing.T) {
	list := []string{
		"",
		"child",
		"inherited",
	}

	for _, s := range list {
		var pe struct {
			Pod userconfig.PodEnum
		}
		data := fmt.Sprintf(`{"pod": "%s"}`, s)
		if err := json.Unmarshal([]byte(data), &pe); err == nil {
			t.Fatalf("Invalid pod value '%s' considered valid", s)
		}
	}
}
