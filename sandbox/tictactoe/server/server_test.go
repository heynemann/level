// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package tictactoe_test

import (
	"time"

	"github.com/heynemann/level/sandbox/tictactoe/server"
	. "github.com/heynemann/level/testing"
	gnatsdServer "github.com/nats-io/gnatsd/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var _ = Describe("TicTacToeServer", func() {
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

	Describe("Connecting to the Game", func() {
		It("should send and receive heartbeat", func() {
			s := &tictactoe.GameplayService{
				ServiceID: uuid.NewV4().String(),
			}
			channel, service, err := RunService(7575, s, logger, "../../../config/test.yaml")
			Expect(err).NotTo(HaveOccurred())
			defer channel.Close()
			defer service.Close()

			conn, err := NewChannelTestConnection(channel)
			defer conn.Close()
			Expect(err).NotTo(HaveOccurred())

			conn.WaitFor(1)
			Expect(conn).To(HaveEvent("channel.session.started"))
			Expect(conn.Received).To(HaveLen(1))
			Expect(conn.Received[0]).To(HavePayload("sessionID"))

			Expect(conn.SessionID).NotTo(BeNil())
		})

		It("should send start game and receive match", func() {
			s := &tictactoe.GameplayService{
				ServiceID: uuid.NewV4().String(),
			}
			channel, service, err := RunService(7575, s, logger, "../../../config/test.yaml")
			Expect(err).NotTo(HaveOccurred())
			defer channel.Close()
			defer service.Close()

			conn, err := NewChannelTestConnection(channel)
			defer conn.Close()
			Expect(err).NotTo(HaveOccurred())

			conn.WaitFor(1)
			Expect(conn).To(HaveEvent("channel.session.started"))

			conn.SendAction(
				"tictactoe.gameplay.start", map[string]interface{}{},
			)

			conn.WaitFor(1)
			time.Sleep(5 * time.Millisecond)
			Expect(conn).To(HaveEvent("tictactoe.gameplay.started"))
			Expect(conn.Received).To(HaveLen(2))
			Expect(conn.Received[1]).To(HavePayload("gameID"))
			Expect(conn.Received[1]).To(HavePayloadLike("opponent", "bot"))
		})

		It("should play with bot", func() {
			s := &tictactoe.GameplayService{
				RandomSeed: 12345678,
				ServiceID:  uuid.NewV4().String(),
			}
			channel, service, err := RunService(7575, s, logger, "../../../config/test.yaml")
			Expect(err).NotTo(HaveOccurred())
			defer channel.Close()
			defer service.Close()

			conn, err := NewChannelTestConnection(channel)
			defer conn.Close()
			Expect(err).NotTo(HaveOccurred())

			conn.WaitFor(1)
			Expect(conn).To(HaveEvent("channel.session.started"))

			conn.SendAction("tictactoe.gameplay.start", map[string]interface{}{})

			conn.WaitFor(1)
			time.Sleep(5 * time.Millisecond)
			Expect(conn).To(HaveEvent("tictactoe.gameplay.started"))

			gameData := conn.Received[len(conn.Received)-1].Payload.(map[string]interface{})
			gameID := gameData["gameID"].(string)

			conn.SendAction("tictactoe.gameplay.move", map[string]interface{}{
				"gameID": gameID,
				"posY":   1,
				"posX":   1,
			})

			conn.WaitFor(1)
			time.Sleep(5 * time.Millisecond)
			Expect(conn).To(HaveEvent("tictactoe.gameplay.status"))
			Expect(conn.Received).To(HaveLen(3))
			Expect(conn.Received[1]).To(HavePayload("gameID"))
			Expect(conn.Received[1]).To(HavePayload("board"))

			gameData = conn.Received[len(conn.Received)-1].Payload.(map[string]interface{})
			board := gameData["board"].([][]int)
			Expect(board[1][1]).To(Equal(1))

			for i := 0; i < 3; i++ {
				for j := 0; i < 3; i++ {
					Expect(board[i][j]).To(Equal(0))
				}
			}
		})
	})
})
