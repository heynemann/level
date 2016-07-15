// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package pubsub_test

import (
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/net/websocket"

	"github.com/heynemann/level/extensions/pubsub"
	"github.com/heynemann/level/extensions/sessionManager"
	"github.com/heynemann/level/messaging"
	"github.com/heynemann/level/services/heartbeat"
	. "github.com/heynemann/level/testing"
	gnatsServer "github.com/nats-io/gnatsd/server"
	"github.com/nats-io/nats"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var defaultDuration = 5 * time.Second

var _ = Describe("Pubsub", func() {

	var logger *MockLogger
	var NATSServer *gnatsServer.Server
	var manager sessionManager.SessionManager
	BeforeEach(func() {
		var err error

		logger = NewMockLogger()
		NATSServer = RunDefaultServer()
		manager, err = sessionManager.GetRedisSessionManager("0.0.0.0", 7777, "", 0, 180, logger)
		Expect(err).NotTo(HaveOccurred())
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
				pubSub, err := pubsub.New(nats.DefaultURL, logger, manager, defaultDuration)
				Expect(err).NotTo(HaveOccurred())

				pubSub.SubscribeActions(serverName, func(reply func(*messaging.Event), action *messaging.Action) {
					receivedAction = action
					reply(messaging.NewEvent("some-event", map[string]interface{}{"x": 2}))
				})

				expectedAction := messaging.NewAction("", "some-action", map[string]interface{}{"a": 1})
				var event *messaging.Event
				err = pubSub.RequestAction(expectedAction, func(ev *messaging.Event) error {
					event = ev
					return nil
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(event).To(BeNil())
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
				pubSub, err := pubsub.New(nats.DefaultURL, logger, manager, defaultDuration)
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

	Describe("WebSocket connection", func() {
		It("Should allow websocket connection", func() {
			pubSub, err := pubsub.New(nats.DefaultURL, logger, manager, defaultDuration)
			Expect(err).NotTo(HaveOccurred())

			port, server, responses := getServer(pubSub)
			Expect(server).NotTo(BeNil())
			time.Sleep(100 * time.Millisecond)

			wsURL := fmt.Sprintf("ws://localhost:%d", port)
			httpURL := fmt.Sprintf("http://localhost:%d", port)
			_, err = websocket.Dial(wsURL, "", httpURL)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(10 * time.Millisecond)

			Expect(pubSub.ConnectedPlayers).To(HaveLen(1))
			Expect(*responses).To(HaveLen(0))
		})

		It("Can send heartbeat", func() {
			hb := heartbeat.NewHeartbeatService()
			pubSub, err := pubsub.New(nats.DefaultURL, logger, manager, defaultDuration, hb)
			Expect(err).NotTo(HaveOccurred())

			port, server, responses := getServer(pubSub)
			Expect(server).NotTo(BeNil())
			time.Sleep(100 * time.Millisecond)

			wsURL := fmt.Sprintf("ws://localhost:%d", port)
			httpURL := fmt.Sprintf("http://localhost:%d", port)
			ws, err := websocket.Dial(wsURL, "", httpURL)

			Expect(err).NotTo(HaveOccurred())

			time.Sleep(10 * time.Millisecond)

			Expect(pubSub.ConnectedPlayers).To(HaveLen(1))

			ping, _ := json.Marshal(map[string]interface{}{
				"type": "action",
				"key":  "channel.heartbeat",
				"payload": map[string]interface{}{
					"clientSent": time.Now().UnixNano() / 1000000,
				},
			})

			_, err = ws.Write(ping)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(10 * time.Millisecond)

			Expect(*responses).To(HaveLen(1))

			var resp map[string]interface{}
			err = json.Unmarshal([]byte((*responses)[0]), &resp)
			Expect(err).NotTo(HaveOccurred())

			Expect(resp["type"]).To(Equal("action"))
			Expect(resp["key"]).To(Equal("channel.heartbeat"))

			Expect(resp["payload"].(map[string]interface{})["clientSent"]).To(BeNumerically(">", 0))

			ws.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			var msg = make([]byte, 512)
			n, err := ws.Read(msg)
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(500))
		})
	})
})
