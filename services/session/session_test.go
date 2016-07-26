// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package session_test

import (
	"github.com/heynemann/level/extensions/serviceRegistry"
	"github.com/heynemann/level/messaging"
	"github.com/heynemann/level/services/session"
	. "github.com/heynemann/level/testing"
	gnatsServer "github.com/nats-io/gnatsd/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var _ = Describe("Session Service", func() {
	var logger *MockLogger
	var NATSServer *gnatsServer.Server
	var reg *registry.ServiceRegistry

	BeforeEach(func() {
		var err error

		logger = NewMockLogger()
		NATSServer = RunServerOnPort(5555)
		reg, err = registry.NewServiceRegistry("nats://127.0.0.1:5555", logger)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		reg.Terminate()
		NATSServer.Shutdown()
		NATSServer = nil
	})

	It("Should handle session action", func() {
		service, err := session.NewSessionService(uuid.NewV4().String(), reg)
		Expect(err).NotTo(HaveOccurred())

		sessionID := uuid.NewV4().String()

		action := messaging.NewAction(
			sessionID,
			"channel.session.start",
			nil,
		)

		Expect(action).NotTo(BeNil())

		event, err := service.HandleAction("channel.session", action)
		Expect(err).NotTo(HaveOccurred())

		Expect(event.Key).To(Equal("channel.session.started"))
		Expect(event.Payload).NotTo(BeNil())

		p := event.Payload.(map[string]interface{})
		Expect(p["sessionID"]).To(Equal(sessionID))
	})
})
