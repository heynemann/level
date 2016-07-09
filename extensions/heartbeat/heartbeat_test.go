// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package heartbeat_test

import (
	"fmt"
	"strconv"
	"time"

	"gopkg.in/redis.v4"

	"github.com/heynemann/level/extensions/heartbeat"
	. "github.com/heynemann/level/testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/uber-go/zap"
)

func verifyServerRegistered(cli *redis.Client, serverName string) {
	serverKey := fmt.Sprintf("server-status:%s", serverName)
	result, err := cli.Get(serverKey).Result()
	Expect(err).NotTo(HaveOccurred())
	value, err := strconv.ParseInt(result, 10, 64)
	Expect(err).NotTo(HaveOccurred())
	Expect(int(value)).To(BeNumerically(">", 0))

	list, err := cli.SMembers("available-servers").Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(list[0]).To(Equal(serverName))
}

var _ = Describe("Heartbeat", func() {
	var testClient *redis.Client
	var logger *MockLogger

	BeforeEach(func() {
		logger = NewMockLogger()
		testClient = redis.NewClient(&redis.Options{
			Addr:     "localhost:7777",
			Password: "",
			DB:       0,
		})

		_, err := testClient.FlushAll().Result()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Heartbeat Extension", func() {
		Describe("Heartbeat creation", func() {
			It("should create new heartbeat", func() {
				heartbeat, err := heartbeat.NewDefault("other-server", "localhost", 7777, "", 0, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(heartbeat).NotTo(BeNil())
				Expect(logger.Messages).To(HaveLen(2))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Connecting to Redis...",
					"source", "heartbeat",
					"redisHost", "localhost",
					"redisPort", 7777,
					"redisDB", 0,
				))

				Expect(logger).To(HaveLogMessage(
					zap.InfoLevel, "Connected to Redis successfully.",
					"source", "heartbeat",
					"redisHost", "localhost",
					"redisPort", 7777,
					"redisDB", 0,
				))
			})
		})

		Describe("Heartbeat registry", func() {
			It("should register server in redis", func() {
				heartbeat, err := heartbeat.NewDefault("other-server", "localhost", 7777, "", 0, logger)
				Expect(err).NotTo(HaveOccurred())

				err = heartbeat.Register()
				Expect(err).NotTo(HaveOccurred())

				verifyServerRegistered(testClient, "other-server")
			})
		})

		Describe("Heartbeat Started", func() {
			It("should register server with start", func() {
				heartbeat, err := heartbeat.New("some-other-server", "localhost", 7777, "", 0, logger, 10*time.Second, 10*time.Millisecond)
				Expect(err).NotTo(HaveOccurred())

				done := heartbeat.Start()
				time.Sleep(time.Millisecond)
				defer func(close chan bool) {
					close <- true
				}(done)

				Expect(err).NotTo(HaveOccurred())
				verifyServerRegistered(testClient, "some-other-server")
			})
		})
	})
})
