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
})
