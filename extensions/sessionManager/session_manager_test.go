// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package sessionManager_test

import (
	"fmt"

	redisCli "gopkg.in/redis.v4"

	"github.com/heynemann/level/extensions/sessionManager"
	"github.com/satori/go.uuid"
	"github.com/uber-go/zap"

	. "github.com/heynemann/level/testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getFaultyRedisClient() *redisCli.Client {
	return redisCli.NewClient(&redisCli.Options{
		Addr:     "0.0.0.0:9876",
		Password: "",
		DB:       0,
	})
}

var _ = Describe("Session Management", func() {

	var testClient *redisCli.Client
	var logger *MockLogger

	BeforeEach(func() {
		logger = NewMockLogger()
		testClient = redisCli.NewClient(&redisCli.Options{
			Addr:     "localhost:7777",
			Password: "",
			DB:       0,
		})
	})

	Describe("Session Manager", func() {

		Describe("Initialized properly", func() {

			It("When getting a new instance", func() {
				sessionManager, err := sessionManager.GetSessionManager(
					"localhost", // Redis Host
					7777,        // Redis Port
					"",          // Redis Pass
					0,           // Redis DB
					180,
					logger,
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(sessionManager).NotTo(BeNil())
				Expect(sessionManager.Logger).NotTo(BeNil())
				Expect(sessionManager.Client).NotTo(BeNil())
			})
		})

		Describe("Initialized with wrong params", func() {
			It("should not be connected to Redis", func() {
				sessionManager, err := sessionManager.GetSessionManager(
					"localhost", // Redis Host
					1249,        // Redis Port
					"",          // Redis Pass
					0,           // Redis DB
					180,
					logger,
				)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("connection refused"))
				Expect(sessionManager).To(BeNil())
			})
		})

		Describe("can start sessions", func() {
			It("should start a session when provided with session id", func() {
				sm := getDefaultSM(logger)
				sessionID := uuid.NewV4().String()
				err := sm.Start(sessionID)
				Expect(err).NotTo(HaveOccurred())

				hashKey := fmt.Sprintf("session-%s", sessionID)
				exists, err := testClient.Exists(hashKey).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())

				lastUpdated, err := testClient.HGet(hashKey, sessionManager.GetLastUpdatedKey()).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(lastUpdated).NotTo(BeNil())

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Starting new session.",
					"source", "sessionManager",
					"operation", "Start",
				))

				Expect(logger).To(HaveLogMessage(
					zap.InfoLevel, "Started session successfully.",
					"source", "sessionManager",
					"operation", "Start",
				))
			})

			It("should fail to start a session when invalid connection", func() {
				sm := getDefaultSM(logger)
				sessionID := uuid.NewV4().String()

				sm.Client = getFaultyRedisClient()
				err := sm.Start(sessionID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("connection refused"))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Starting new session.",
					"source", "sessionManager",
					"operation", "Start",
				))

				Expect(logger).To(HaveLogMessage(
					zap.ErrorLevel, "Could not start session.",
					"source", "sessionManager",
					"operation", "Start",
					"error", "dial tcp 0.0.0.0:9876: getsockopt: connection refused",
				))

			})
		})

		Describe("can merge sessions", func() {
			It("should merge a session into another one", func() {
				sm := getDefaultSM(logger)

				oldSessionID := uuid.NewV4().String()
				sm.Start(oldSessionID)
				hashKey := fmt.Sprintf("session-%s", oldSessionID)
				testClient.HSet(hashKey, "someKey", "someValue")

				sessionID := uuid.NewV4().String()
				sm.Start(sessionID)
				count, err := sm.Merge(oldSessionID, sessionID)
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(1))

				exists, err := testClient.Exists(hashKey).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeFalse())

				hashKey = fmt.Sprintf("session-%s", sessionID)
				exists, err = testClient.Exists(hashKey).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())

				someValue, err := testClient.HGet(hashKey, "someKey").Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(someValue).To(Equal("someValue"))
			})

			It("should not merge a non-existing session", func() {
				sm := getDefaultSM(logger)

				sessionID := uuid.NewV4().String()
				sm.Start(sessionID)

				count, err := sm.Merge("invalid-id", sessionID)
				Expect(count).To(Equal(0))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Session with session ID invalid-id was not found in session storage."))
			})
		})

		Describe("can get session", func() {
			It("should be able to load a session", func() {
				sm := getDefaultSM(logger)
				sessionID := uuid.NewV4().String()
				sm.Start(sessionID)

				session, err := sm.Load(sessionID)

				Expect(err).ToNot(HaveOccurred())

				Expect(session.ID).To(Equal(sessionID))
				Expect(session.Manager).To(BeEquivalentTo(sm))
				Expect(session.Get(sessionManager.GetLastUpdatedKey())).To(BeNumerically(">", 0))
			})

			It("should not load a session if invalid id", func() {
				sm := getDefaultSM(logger)
				sessionID := uuid.NewV4().String()

				session, err := sm.Load(sessionID)
				Expect(session).To(BeNil())
				Expect(err).To(HaveOccurred())
				expected := fmt.Sprintf("Session with session ID %s was not found in session storage.", sessionID)
				Expect(err.Error()).To(Equal(expected))
			})
		})
	})
})
