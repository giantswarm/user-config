package userconfig_test

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	userConfigPkg "github.com/giantswarm/user-config"
)

func TestUserConfigValidator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "user config validator")
}

var _ = Describe("user config validator", func() {
	var (
		err       error
		byteSlice []byte
		appConfig userConfigPkg.AppConfig
	)

	BeforeEach(func() {
		err = nil
	})

	Describe("UnmarshalSwarmJson()", func() {
		Describe("parsing valid app-config", func() {
			BeforeEach(func() {
				byteSlice = []byte(`{
          "app_name": "test-app-name"
        }`)

				appConfig = userConfigPkg.AppConfig{}
				err = userConfigPkg.UnmarshalSwarmJson(byteSlice, &appConfig)
			})

			It("should not throw error", func() {
				Expect(err).To(BeNil())
			})

			It("should properly parse given app name", func() {
				Expect(appConfig.AppName).To(Equal("test-app-name"))
			})
		})

		Describe("parsing app-config with unknown fields", func() {
			BeforeEach(func() {
				byteSlice = []byte(`{
          "app_name": "test-app-name",
          "foo": 47
        }`)

				appConfig = userConfigPkg.AppConfig{}
				err = userConfigPkg.UnmarshalSwarmJson(byteSlice, &appConfig)
			})

			It("should throw error ErrUnknownJSONField", func() {
				Expect(userConfigPkg.IsErrUnknownJsonField(err)).To(BeTrue())
			})

			It("should not parse given app name", func() {
				Expect(appConfig.AppName).To(Equal(""))
			})
		})
	})

	Describe("json.Unmarshal()", func() {
		Describe("parsing valid app-config", func() {
			BeforeEach(func() {
				byteSlice = []byte(`{
          "app_name": "test-app-name"
        }`)

				appConfig = userConfigPkg.AppConfig{}
				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should not throw error", func() {
				Expect(err).To(BeNil())
			})

			It("should properly parse given app name", func() {
				Expect(appConfig.AppName).To(Equal("test-app-name"))
			})
		})

		Describe("parsing app-config with unknown fields", func() {
			BeforeEach(func() {
				byteSlice = []byte(`{
          "app_name": "test-app-name",
          "foo": 47
        }`)

				appConfig = userConfigPkg.AppConfig{}
				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error ErrUnknownJSONField", func() {
				Expect(userConfigPkg.IsErrUnknownJsonField(err)).To(BeTrue())
			})

			It("should not parse given app name", func() {
				Expect(appConfig.AppName).To(Equal(""))
			})
		})
	})
})
