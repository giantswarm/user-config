package userconfig

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestV2UserConfigPodFunctions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "v2 user config pod functions")
}

var _ = Describe("v2 user config pod functions", func() {

	testComponent := func() *ComponentDefinition {
		return &ComponentDefinition{
			Image: MustParseImageDefinition("registry/namespace/repository:version"),
		}
	}

	setPod := func(config *ComponentDefinition, pod PodEnum) *ComponentDefinition {
		config.Pod = pod
		return config
	}

	testApp := func() ComponentDefinitions {
		return make(ComponentDefinitions)
	}

	Describe("pod function tests ", func() {
		Describe("PodComponents with pod=children should return only direct child components with pod!=none", func() {
			var err error
			var list ComponentDefinitions

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = testComponent()
				components["component/a/b2"] = testComponent()
				components["component/a/b2/c"] = testComponent()
				components["component/a/b3"] = setPod(testComponent(), PodNone)
				components["component/b/g1"] = testComponent()

				list, err = components.PodComponents("component/a")
			})

			It("should not throw an error", func() {
				Expect(err).To(BeNil())
			})
			It("list should contain component/a/b1 and component/a/b2", func() {
				Expect(list).To(HaveLen(2))
				Expect(list).To(HaveKey(ComponentName("component/a/b1")))
				Expect(list).To(HaveKey(ComponentName("component/a/b2")))
			})
		})

		Describe("PodComponents with pod=inherit should return all child components with pod!=none", func() {
			var err error
			var list ComponentDefinitions

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodInherit)
				components["component/a/b1"] = testComponent()
				components["component/a/b2"] = testComponent()
				components["component/a/b2/c"] = testComponent()
				components["component/a/b3"] = setPod(testComponent(), PodNone)
				components["component/a/b3/c"] = testComponent()
				components["component/a/b3/c/d"] = testComponent()

				list, err = components.PodComponents("component/a")
			})

			It("should not throw an error", func() {
				Expect(err).To(BeNil())
			})
			It("list should contain component/a/b1 and component/a/b2, component/a/b2/c", func() {
				Expect(list).To(HaveLen(3))
				Expect(list).To(HaveKey(ComponentName("component/a/b1")))
				Expect(list).To(HaveKey(ComponentName("component/a/b2")))
				Expect(list).To(HaveKey(ComponentName("component/a/b2/c")))
			})
		})

		Describe("Calling PodComponents on a component with pod!=children|inherit should return an error", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = testComponent()

				_, err = components.PodComponents("component/a")
			})

			It("should throw an error", func() {
				Expect(IsInvalidArgument(err)).To(BeTrue())
			})
		})

	})
})
