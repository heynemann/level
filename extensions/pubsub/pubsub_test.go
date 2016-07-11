// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package pubsub_test

import (
	"time"

	"github.com/heynemann/level/extensions/pubsub"
	"github.com/heynemann/level/messaging"
	. "github.com/heynemann/level/testing"
	gnatsServer "github.com/nats-io/gnatsd/server"
	"github.com/nats-io/nats"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var _ = Describe("Pubsub", func() {

	var logger *MockLogger
	var NATSServer *gnatsServer.Server
	BeforeEach(func() {
		logger = NewMockLogger()
		NATSServer = RunDefaultServer()
	})
	AfterEach(func() {
		NATSServer.Shutdown()
		NATSServer = nil
	})

	Describe("PubSub Extension", func() {
		Describe("Publish/Subscribe Actions", func() {
			It("should allow servers to subscribe to actions and for them to be published", func() {
				var receivedAction *messaging.Action
				serverName := uuid.NewV4().String()
				pubSub, err := pubsub.New(nats.DefaultURL, logger)
				Expect(err).NotTo(HaveOccurred())

				pubSub.SubscribeActions(serverName, func(reply func(*messaging.Event), action *messaging.Action) {
					receivedAction = action
					reply(messaging.NewEvent("some-event", map[string]interface{}{"x": 2}))
				})

				expectedAction := messaging.NewAction("some-action", map[string]interface{}{"a": 1})
				event, err := pubSub.RequestAction(serverName, expectedAction, 10*time.Millisecond)
				Expect(err).NotTo(HaveOccurred())
				Expect(event).NotTo(BeNil())
				Expect(event.Type).To(Equal("some-event"))
				Expect(event.Payload).To(MapEqual(map[string]interface{}{"x": 2}))

				time.Sleep(time.Millisecond)

				Expect(receivedAction).NotTo(BeNil())
				Expect(receivedAction.Type).To(Equal(expectedAction.Type))
				Expect(receivedAction.Payload).To(MapEqual(expectedAction.Payload))
			})
		})

		Describe("Publish/Subscribe Messages", func() {
			It("should allow servers to publish events to clients", func() {
				var receivedEvent *messaging.Event
				pubSub, err := pubsub.New(nats.DefaultURL, logger)
				Expect(err).NotTo(HaveOccurred())

				pubSub.SubscribeEvents(func(event *messaging.Event) {
					receivedEvent = event
				})

				expectedEvent := messaging.NewEvent("some-event", map[string]interface{}{"a": 1})
				pubSub.PublishEvent(expectedEvent)

				time.Sleep(time.Millisecond)

				Expect(receivedEvent).NotTo(BeNil())
				Expect(receivedEvent.Type).To(Equal(expectedEvent.Type))
				Expect(receivedEvent.Payload).To(MapEqual(expectedEvent.Payload))
			})
		})
	})
})
