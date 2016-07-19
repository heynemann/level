// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package heartbeat_test

import (
	"time"

	"github.com/heynemann/level/extensions/serviceRegistry"
	"github.com/heynemann/level/messaging"
	"github.com/heynemann/level/services/heartbeat"
	. "github.com/heynemann/level/testing"
	gnatsServer "github.com/nats-io/gnatsd/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var _ = Describe("Heartbeat Service", func() {
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

	It("Should handle heartbeat action", func() {
		service, err := heartbeat.NewHeartbeatService(uuid.NewV4().String(), reg)
		Expect(err).NotTo(HaveOccurred())

		action := messaging.NewAction(
			uuid.NewV4().String(),
			"channel.heartbeat",
			map[string]interface{}{
				"clientSent": time.Now().UnixNano(),
			},
		)

		Expect(action).NotTo(BeNil())

		event, err := service.HandleAction("", action)
		Expect(err).NotTo(HaveOccurred())
		Expect(event).NotTo(BeNil())
		Expect(event.Payload.(map[string]interface{})["serverSent"]).To(BeNumerically(">", 0))
	})

	It("Should fail in case of wrong message", func() {
		service, err := heartbeat.NewHeartbeatService(uuid.NewV4().String(), reg)
		Expect(err).NotTo(HaveOccurred())

		action := messaging.NewAction(
			uuid.NewV4().String(),
			"channel.heartbeat",
			"invalid-payload",
		)

		Expect(action).NotTo(BeNil())

		_, err = service.HandleAction("subject", action)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Could not understand heartbeat payload: invalid-payload"))
	})
})
