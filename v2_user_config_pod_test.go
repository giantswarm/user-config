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

	testNode := func() *NodeDefinition {
		return &NodeDefinition{
			Image: MustParseImageDefinition("registry/namespace/repository:version"),
		}
	}

	setPod := func(config *NodeDefinition, pod PodEnum) *NodeDefinition {
		config.Pod = pod
		return config
	}

	testApp := func() NodeDefinitions {
		return make(NodeDefinitions)
	}

	Describe("pod function tests ", func() {
		Describe("PodNodes with pod=children should return only direct child nodes with pod!=none", func() {
			var err error
			var list NodeDefinitions

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = testNode()
				nodes["node/a/b2"] = testNode()
				nodes["node/a/b2/c"] = testNode()
				nodes["node/a/b3"] = setPod(testNode(), PodNone)
				nodes["node/b/g1"] = testNode()

				list, err = nodes.PodNodes("node/a")
			})

			It("should not throw an error", func() {
				Expect(err).To(BeNil())
			})
			It("list should contain node/a/b1 and node/a/b2", func() {
				Expect(list).To(HaveLen(2))
				Expect(list).To(HaveKey(NodeName("node/a/b1")))
				Expect(list).To(HaveKey(NodeName("node/a/b2")))
			})
		})

		Describe("PodNodes with pod=inherit should return all child nodes with pod!=none", func() {
			var err error
			var list NodeDefinitions

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodInherit)
				nodes["node/a/b1"] = testNode()
				nodes["node/a/b2"] = testNode()
				nodes["node/a/b2/c"] = testNode()
				nodes["node/a/b3"] = setPod(testNode(), PodNone)
				nodes["node/a/b3/c"] = testNode()
				nodes["node/a/b3/c/d"] = testNode()

				list, err = nodes.PodNodes("node/a")
			})

			It("should not throw an error", func() {
				Expect(err).To(BeNil())
			})
			It("list should contain node/a/b1 and node/a/b2, node/a/b2/c", func() {
				Expect(list).To(HaveLen(3))
				Expect(list).To(HaveKey(NodeName("node/a/b1")))
				Expect(list).To(HaveKey(NodeName("node/a/b2")))
				Expect(list).To(HaveKey(NodeName("node/a/b2/c")))
			})
		})

		Describe("Calling PodNodes on a node with pod!=children|inherit should return an error", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = testNode()

				_, err = nodes.PodNodes("node/a")
			})

			It("should throw an error", func() {
				Expect(IsInvalidArgument(err)).To(BeTrue())
			})
		})

	})
})
