package chroma_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	chroma "github.com/urjitbhatia/gochroma"
)

var _ = Describe("Client", func() {
	Describe("Connection", func() {
		It("connects properly", func() {
			_, err := chroma.NewClient("http://localhost:8000")
			Expect(err).ToNot(HaveOccurred())
		})

		It("rejects bad server urls", func() {
			_, err := chroma.NewClient("http\n://foo.com/")
			Expect(err).To(HaveOccurred())
		})
	})

	It("gets heartbeat", func() {
		client, err := chroma.NewClient("http://localhost:8000")
		Expect(err).ToNot(HaveOccurred())

		alive, err := client.Heartbeat()
		Expect(err).ToNot(HaveOccurred())
		Expect(alive).To(BeNumerically(">", 0))

	})

	It("resets the db", func() {
		client, err := chroma.NewClient("http://localhost:8000")
		Expect(err).ToNot(HaveOccurred())

		ok, err := client.Reset()
		Expect(err).ToNot(HaveOccurred())
		Expect(ok).To(BeTrue())
	})

	It("gets the version of the db", func() {
		client, err := chroma.NewClient("http://localhost:8000")
		Expect(err).ToNot(HaveOccurred())

		ver, err := client.GetVersion()
		Expect(err).ToNot(HaveOccurred())
		Expect(ver).To(Equal("0.4.8"))
	})

	Describe("collections", func() {
		BeforeEach(func() {
			client, err := chroma.NewClient("http://localhost:8000")
			Expect(err).ToNot(HaveOccurred())
			client.Reset()
		})

		It("list collections", func() {
			client, err := chroma.NewClient("http://localhost:8000")
			Expect(err).ToNot(HaveOccurred())

			collections, err := client.ListCollections()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(collections)).To(Equal(0))
		})

		It("create and then list collection", func() {
			client, err := chroma.NewClient("http://localhost:8000")
			Expect(err).ToNot(HaveOccurred())

			// create new collection
			collection, err := client.CreateCollection("unit-test", "l2", nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(collection.Name).To(Equal("unit-test"))

			// should error if recreating existing collection
			_, err = client.CreateCollection("unit-test", "l2", nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error while creating collection: ValueError('Collection unit-test already exists.')"))

			// list the collections
			collections, err := client.ListCollections()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(collections)).To(Equal(1))
			Expect(collections[0].Name).To(Equal("unit-test"))
			Expect(collections[0].Metadata).To(Equal(map[string]any{"hnsw:space": "l2"}))
		})

		It("delete and then list collections", func() {
			client, err := chroma.NewClient("http://localhost:8000")
			Expect(err).ToNot(HaveOccurred())

			// delete when collection doesn't exist should error
			err = client.DeleteCollection("unittest")
			Expect(err.Error()).To(Equal("error deleting collection: ValueError('Collection unittest does not exist.')"))

			// create a collection, then delete
			_, err = client.CreateCollection("unittest", "l2", nil)
			Expect(err).ToNot(HaveOccurred())

			// this time delete shouldn't error
			err = client.DeleteCollection("unittest")
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
