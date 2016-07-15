// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package heartbeat_test

import (
	"time"

	"github.com/heynemann/level/messaging"
	"github.com/heynemann/level/services/heartbeat"
	. "github.com/heynemann/level/testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var _ = Describe("Heartbeat Service", func() {
	var logger *MockLogger

	BeforeEach(func() {
		logger = NewMockLogger()
	})

	It("Should handle heartbeat action", func() {
		service := heartbeat.NewHeartbeatService()

		action := messaging.NewAction(
			uuid.NewV4().String(),
			"channel.heartbeat",
			map[string]interface{}{
				"clientStart": time.Now().UnixNano(),
				"serverStart": time.Now().UnixNano(),
			},
		)

		Expect(action).NotTo(BeNil())

		service.HandleAction(action)
	})
})
