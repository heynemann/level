// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package pubsub_test

import (
	"fmt"
	"time"

	"golang.org/x/net/websocket"

	"github.com/heynemann/level/extensions/pubsub"
	"github.com/heynemann/level/extensions/sessionManager"
	. "github.com/heynemann/level/testing"
	gnatsServer "github.com/nats-io/gnatsd/server"
	"github.com/nats-io/nats"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

			//It("Can send heartbeat", func() {
			//hb := heartbeat.NewHeartbeatService()
			//pubSub, err := pubsub.New(nats.DefaultURL, logger, manager, defaultDuration, hb)
			//Expect(err).NotTo(HaveOccurred())

			//port, server, responses := getServer(pubSub)
			//Expect(server).NotTo(BeNil())
			//time.Sleep(100 * time.Millisecond)

			//wsURL := fmt.Sprintf("ws://localhost:%d", port)
			//httpURL := fmt.Sprintf("http://localhost:%d", port)
			//ws, err := websocket.Dial(wsURL, "", httpURL)

			//Expect(err).NotTo(HaveOccurred())

			//time.Sleep(10 * time.Millisecond)

			//Expect(pubSub.ConnectedPlayers).To(HaveLen(1))

			//ping := messaging.NewAction(
			//"",
			//"channel.heartbeat",
			//map[string]interface{}{
			//"clientSent": time.Now().UnixNano() / 1000000,
			//},
			//)
			//pingJSON, _ := ping.MarshalJSON()
			//_, err = ws.Write(pingJSON)
			//Expect(err).NotTo(HaveOccurred())

			//time.Sleep(10 * time.Millisecond)

			//Expect(*responses).To(HaveLen(1))

			//receivedAction := messaging.Action{}
			//err = receivedAction.UnmarshalJSON([]byte((*responses)[0]))
			//Expect(err).NotTo(HaveOccurred())

			//Expect(receivedAction.Key).To(Equal("channel.heartbeat"))
			//p := receivedAction.Payload.(map[string]interface{})
			//Expect(p["clientSent"]).To(BeNumerically(">", 0))

			//ws.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			//var msg = make([]byte, 1024)
			//_, err = ws.Read(msg)
			//Expect(err).NotTo(HaveOccurred())

			//event := messaging.Event{}
			//event.UnmarshalJSON(msg)

			//Expect(event.Key).To(Equal("channel.heartbeat"))
			//p = event.Payload.(map[string]interface{})
			//Expect(p["serverReceived"]).To(BeNumerically(">", 0))
			//Expect(p["serverSent"]).To(BeNumerically(">", 0))
			//})
		})
	})
})
