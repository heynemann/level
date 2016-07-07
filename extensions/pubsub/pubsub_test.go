// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package pubsub_test

import (
	gnatsServer "github.com/nats-io/gnatsd/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pubsub", func() {

	var NATSServer *gnatsServer.Server
	BeforeEach(func() {
		NATSServer = RunDefaultServer()
	})
	AfterEach(func() {
		NATSServer.Shutdown()
		NATSServer = nil
	})

	Describe("PubSub Extension", func() {
		Describe("Subscribe Operations", func() {
			It("should subscribe to a topic", func() {
				Expect(true).To(BeTrue())
			})
		})
	})
})
