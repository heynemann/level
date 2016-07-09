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

func getFaultyRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "0.0.0.0:9876",
		Password: "",
		DB:       0,
	})
}

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
				))

				Expect(logger).To(HaveLogMessage(
					zap.InfoLevel, "Connected to Redis successfully.",
					"source", "heartbeat",
				))
			})

			It("should not create a heartbeat if redis connection is wrong", func() {
				heartbeat, err := heartbeat.NewDefault("other-server", "localhost", 8987, "", 0, logger)
				Expect(heartbeat).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("connection refused"))

				Expect(logger).To(HaveLogMessage(
					zap.ErrorLevel, "Could not connect to redis.",
					"source", "heartbeat",
					"error", "dial tcp 127.0.0.1:8987: getsockopt: connection refused",
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

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Registering server with service registry...",
					"source", "heartbeat",
				))

				Expect(logger).To(HaveLogMessage(
					zap.InfoLevel, "Registered with service registry successfully.",
					"source", "heartbeat",
				))
			})

			It("should fail register server in redis when pass is wrong", func() {
				heartbeat, err := heartbeat.NewDefault("other-server", "localhost", 7777, "", 0, logger)
				Expect(err).NotTo(HaveOccurred())

				heartbeat.Client = getFaultyRedisClient()
				err = heartbeat.Register()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("connection refused"))

				Expect(logger).To(HaveLogMessage(
					zap.ErrorLevel, "Could not register with service registry.",
					"source", "heartbeat",
					"error", "dial tcp 0.0.0.0:9876: getsockopt: connection refused",
				))
			})

		})

		Describe("Heartbeat Started", func() {
			It("should register server with start", func() {
				heartbeat, err := heartbeat.New("some-other-server", "localhost", 7777, "", 0, logger, 10*time.Second, 10*time.Millisecond)
				Expect(err).NotTo(HaveOccurred())

				done := heartbeat.Start()
				time.Sleep(time.Millisecond)
				done <- true

				Expect(err).NotTo(HaveOccurred())
				verifyServerRegistered(testClient, "some-other-server")

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Starting heartbeat...",
					"source", "heartbeat",
				))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Stopping heartbeat...",
					"source", "heartbeat",
				))

				Expect(logger).To(HaveLogMessage(
					zap.InfoLevel, "Status updated successfully in redis.",
					"source", "heartbeat",
				))
			})
		})
	})
})
