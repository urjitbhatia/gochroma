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
})
