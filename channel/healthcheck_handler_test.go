// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package channel_test

import (
	"net/http"

	. "github.com/heynemann/level/testing"
	gnatsdServer "github.com/nats-io/gnatsd/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Healthcheck Handler", func() {
	var logger *MockLogger
	var NATSServer *gnatsdServer.Server

	BeforeEach(func() {
		logger = NewMockLogger()
		NATSServer = RunServerOnPort(7778)
	})
	AfterEach(func() {
		NATSServer.Shutdown()
		NATSServer = nil
	})

	XIt("Should respond with default WORKING string", func() {
		a, err := GetDefaultTestApp(logger)
		Expect(err).NotTo(HaveOccurred())
		res := Get(a, "/healthcheck")

		Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
		Expect(res.Body().Raw()).To(Equal("WORKING"))
	})

	XIt("Should respond with customized WORKING string", func() {
		a, err := GetDefaultTestApp(logger)
		Expect(err).NotTo(HaveOccurred())

		a.Config.Set("channel.workingText", "OTHERWORKING")
		res := Get(a, "/healthcheck")

		Expect(res.Raw().StatusCode).To(Equal(http.StatusOK))
		Expect(res.Body().Raw()).To(Equal("OTHERWORKING"))
	})
})
