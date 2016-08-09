// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package tictactoe_test

import (
	"fmt"
	"math/rand"

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
			s := tictactoe.NewGameplayService(uuid.NewV4().String())
			channel, service, conn, err := GetTestClient(7575, s, logger, "../../../config/test.yaml")
			Expect(err).NotTo(HaveOccurred())
			defer channel.Close()
			defer service.Close()
			defer conn.Close()

			ev, err := conn.WaitForEvent("channel.session.started")
			Expect(err).NotTo(HaveOccurred())
			Expect(ev).To(HavePayload("sessionID"))

			Expect(conn.SessionID).NotTo(BeNil())
		})

		It("should send start game and receive match", func() {
			s := tictactoe.NewGameplayService(uuid.NewV4().String())
			channel, service, conn, err := GetTestClient(7575, s, logger, "../../../config/test.yaml")
			Expect(err).NotTo(HaveOccurred())
			defer channel.Close()
			defer service.Close()
			defer conn.Close()

			ev, err := conn.WaitForEvent("channel.session.started")
			Expect(err).NotTo(HaveOccurred())

			conn.SendAction(
				"tictactoe.gameplay.start", map[string]interface{}{},
			)
			ev, err = conn.WaitForEvent("tictactoe.gameplay.started")
			Expect(err).NotTo(HaveOccurred())
			Expect(ev).To(HavePayload("gameID"))
			Expect(ev).To(HavePayloadLike("opponent", "bot"))
		})

		It("should play with bot", func() {
			rand.Seed(12345678)
			s := tictactoe.NewGameplayService(uuid.NewV4().String())
			channel, service, conn, err := GetTestClient(7575, s, logger, "../../../config/test.yaml")
			Expect(err).NotTo(HaveOccurred())
			defer channel.Close()
			defer service.Close()
			defer conn.Close()

			conn.WaitForEvent("channel.session.started")
			conn.SendAction("tictactoe.gameplay.start", map[string]interface{}{})
			ev, err := conn.WaitForEvent("tictactoe.gameplay.started")
			Expect(err).NotTo(HaveOccurred())

			gameData := ev.Payload.(map[string]interface{})
			gameID := gameData["gameID"].(string)

			conn.SendAction("tictactoe.gameplay.move", map[string]interface{}{
				"gameID": gameID,
				"posY":   1,
				"posX":   1,
			})

			ev, err = conn.WaitForEvent("tictactoe.gameplay.status")
			Expect(err).NotTo(HaveOccurred())

			Expect(ev).To(HavePayload("gameID"))
			Expect(ev).To(HavePayload("board"))

			gameData = ev.Payload.(map[string]interface{})
			board := gameData["board"].([]interface{})
			Expect(board[1].([]interface{})[1]).To(BeEquivalentTo(1))
			Expect(board[2].([]interface{})[0]).To(BeEquivalentTo(2))

			for i := 0; i < 3; i++ {
				row := board[i].([]interface{})
				for j := 0; j < 3; j++ {
					By(fmt.Sprintf("Comparing X %d, Y %d", i, j))
					if (i == 1 && j == 1) || (i == 2 && j == 0) {
						continue
					}
					Expect(row[j]).To(BeEquivalentTo(0))
				}
			}
		})

		It("should win", func() {
			rand.Seed(12345678)
			s := tictactoe.NewGameplayService(uuid.NewV4().String())
			channel, service, conn, err := GetTestClient(7575, s, logger, "../../../config/test.yaml")
			Expect(err).NotTo(HaveOccurred())
			defer channel.Close()
			defer service.Close()
			defer conn.Close()

			conn.WaitForEvent("channel.session.started")
			conn.SendAction("tictactoe.gameplay.start", map[string]interface{}{})
			ev, err := conn.WaitForEvent("tictactoe.gameplay.started")
			Expect(err).NotTo(HaveOccurred())

			gameData := ev.Payload.(map[string]interface{})
			gameID := gameData["gameID"].(string)

			conn.SendAction("tictactoe.gameplay.move", map[string]interface{}{
				"gameID": gameID,
				"posX":   0,
				"posY":   0,
			})
			_, err = conn.WaitForEvent("tictactoe.gameplay.status")
			Expect(err).NotTo(HaveOccurred())

			conn.SendAction("tictactoe.gameplay.move", map[string]interface{}{
				"gameID": gameID,
				"posX":   0,
				"posY":   1,
			})
			_, err = conn.WaitForEvent("tictactoe.gameplay.status")
			Expect(err).NotTo(HaveOccurred())

			conn.SendAction("tictactoe.gameplay.move", map[string]interface{}{
				"gameID": gameID,
				"posX":   0,
				"posY":   2,
			})

			ev, err = conn.WaitForEvent("tictactoe.gameplay.result")
			Expect(err).NotTo(HaveOccurred())

			Expect(s.Games).To(HaveLen(0))
		})
	})
})
