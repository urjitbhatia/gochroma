package chroma_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	chroma "github.com/urjitbhatia/gochroma"
)

var _ = Describe("Client", func() {
	Describe("Connection", func() {
		It("rejects bad server urls", func() {
			_, err := chroma.NewClient("http\n://foo.com/")
			Expect(err).To(HaveOccurred())
		})
	})

	It("gets heartbeat", func() {
		alive, err := testClient.Heartbeat()
		Expect(err).ToNot(HaveOccurred())
		Expect(alive).To(BeNumerically(">", 0))
	})

	It("resets the db", func() {
		ok, err := testClient.Reset()
		Expect(err).ToNot(HaveOccurred())
		Expect(ok).To(BeTrue())
	})

	It("gets the version of the db", func() {
		ver, err := testClient.GetVersion()
		Expect(err).ToNot(HaveOccurred())
		Expect(ver).To(Equal("0.4.8"))
	})

	Describe("collections", func() {

		Describe("create, list, get collection", Ordered, func() {
			It("reset", func() {
				ok, err := testClient.Reset()
				Expect(ok).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())
			})

			It("create", func() {
				// create new collection
				collection, err := testClient.CreateCollection("unit-test", "l2", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(collection.Name).To(Equal("unit-test"))
			})

			It("create existing", func() {
				// should error if recreating existing collection
				_, err := testClient.CreateCollection("unit-test", "l2", nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("error while creating collection: ValueError('Collection unit-test already exists.')"))
			})

			It("get", func() {
				collection, err := testClient.GetCollection("unit-test")
				Expect(err).ToNot(HaveOccurred())
				Expect(collection.Name).To(Equal("unit-test"))
				Expect(collection.DistanceFn).To(Equal("l2"))
			})

			It("list", func() {
				// list the collections
				collections, err := testClient.ListCollections()
				Expect(err).ToNot(HaveOccurred())
				Expect(len(collections)).To(Equal(1))
				Expect(collections[0].Name).To(Equal("unit-test"))
				Expect(collections[0].Metadata).To(Equal(map[string]any{"hnsw:space": "l2"}))
			})

			It("delete", func() {
				// delete when collection doesn't exist should error
				err := testClient.DeleteCollection("unknown")
				Expect(err.Error()).To(Equal("error deleting collection: ValueError('Collection unknown does not exist.')"))

				// existing collection delete
				err = testClient.DeleteCollection("unit-test")
				Expect(err).ToNot(HaveOccurred())
			})

			It("list after delete", func() {
				// list the collections
				collections, err := testClient.ListCollections()
				Expect(err).ToNot(HaveOccurred())
				Expect(len(collections)).To(Equal(0))
			})

			It("getOrCreate", func() {
				collection, err := testClient.GetOrCreateCollection("unit-test-getorcreate", "l2", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(collection.Name).To(Equal("unit-test-getorcreate"))

				// recreate
				collection, err = testClient.GetOrCreateCollection("unit-test-getorcreate", "l2", nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(collection.Name).To(Equal("unit-test-getorcreate"))
			})
		})

	})
})
