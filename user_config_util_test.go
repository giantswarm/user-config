package userconfig_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	userConfigPkg "github.com/giantswarm/user-config"
)

func TestUserConfigUtil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "user config util")
}

var _ = Describe("user config util", func() {
	Describe("FixFieldName", func() {
		var (
			input    string
			expected string
		)

		It("should transform upper cased character in the middle of the word", func() {
			input = "appName"
			expected = "app_name"
		})

		It("should transform upper cased character at the beginning of the word", func() {
			input = "Services"
			expected = "services"
		})

		It("should transform upper cased character at the beginning and in the middle of the word", func() {
			input = "ComponentName"
			expected = "component_name"
		})

		AfterEach(func() {
			Expect(userConfigPkg.FixFieldName(input)).To(Equal(expected))
		})
	})
})
