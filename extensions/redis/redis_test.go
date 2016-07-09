// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package redis_test

import (
	"github.com/heynemann/level/extensions/redis"
	. "github.com/heynemann/level/testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/uber-go/zap"
)

var _ = Describe("Redis", func() {
	var logger *MockLogger

	BeforeEach(func() {
		logger = NewMockLogger()
	})

	Describe("Redis Extension", func() {
		Describe("Get Redis", func() {
			It("should return proper connection when valid params", func() {
				cli, err := redis.New("localhost", 7777, "", 0, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(cli).NotTo(BeNil())
				Expect(logger.Messages).To(HaveLen(2))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Connecting to Redis...",
					"redisHost", "localhost",
					"redisPort", 7777,
					"redisDB", 0,
				))

				Expect(logger).To(HaveLogMessage(
					zap.InfoLevel, "Connected to Redis successfully.",
					"redisHost", "localhost",
					"redisPort", 7777,
					"redisDB", 0,
				))
			})

			It("should return error when invalid params", func() {
				cli, err := redis.New("localhost", 9876, "", 0, logger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("connection refused"))
				Expect(cli).To(BeNil())
				Expect(logger.Messages).To(HaveLen(2))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Connecting to Redis...",
					"redisHost", "localhost",
					"redisPort", 9876,
					"redisDB", 0,
				))

				Expect(logger).To(HaveLogMessage(
					zap.ErrorLevel, "Could not connect to redis.",
					"redisHost", "localhost",
					"redisPort", 9876,
					"redisDB", 0,
					"error", "dial tcp 127.0.0.1:9876: getsockopt: connection refused",
				))

			})
		})
	})
})
