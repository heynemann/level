// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package heartbeat_test

import (
	"fmt"
	"time"

	"github.com/heynemann/level/extensions/pubsub"
	"github.com/heynemann/level/messaging"
	"github.com/heynemann/level/services/heartbeat"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var _ = Describe("Heartbeat Service", func() {
	It("Should handle heartbeat action", func() {
		service := heartbeat.NewHeartbeatService()
		ps := &pubsub.PubSub{}
		service.Initialize(ps)
		Expect(service.PubSub).To(Equal(ps))

		action := messaging.NewAction(
			uuid.NewV4().String(),
			"channel.heartbeat",
			map[string]interface{}{
				"clientStart": time.Now().UnixNano(),
			},
		)

		Expect(action).NotTo(BeNil())

		called := false
		reply := func(event *messaging.Event) error {
			called = true
			return nil
		}

		service.HandleAction(action, reply, time.Now().UnixNano())
		Expect(called).To(BeTrue())
	})

	It("Should fail in case of wrong message", func() {
		service := heartbeat.NewHeartbeatService()

		action := messaging.NewAction(
			uuid.NewV4().String(),
			"channel.heartbeat",
			"invalid-payload",
		)

		Expect(action).NotTo(BeNil())

		called := false
		reply := func(event *messaging.Event) error {
			called = true
			return nil
		}

		err := service.HandleAction(action, reply, time.Now().UnixNano())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Could not understand heartbeat payload: invalid-payload"))
		Expect(called).To(BeFalse())
	})

	It("Should fail in case of error in reply", func() {
		service := heartbeat.NewHeartbeatService()

		action := messaging.NewAction(
			uuid.NewV4().String(),
			"channel.heartbeat",
			map[string]interface{}{
				"clientStart": time.Now().UnixNano(),
			},
		)

		Expect(action).NotTo(BeNil())

		reply := func(event *messaging.Event) error {
			return fmt.Errorf("failed to reply")
		}

		err := service.HandleAction(action, reply, time.Now().UnixNano())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("failed to reply"))
	})

})
