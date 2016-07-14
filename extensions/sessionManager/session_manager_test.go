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

		Describe("Initialization", func() {

			It("When getting a new instance", func() {
				sessionManager, err := sessionManager.GetRedisSessionManager(
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

			It("should not initialize with wrong arguments", func() {
				sessionManager, err := sessionManager.GetRedisSessionManager(
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

		Describe("Starting sessions", func() {
			It("should start a session when provided with session id", func() {
				sm := getDefaultSM(logger)
				sessionID := uuid.NewV4().String()
				session, err := sm.Start(sessionID)
				Expect(err).NotTo(HaveOccurred())
				Expect(session).NotTo(BeNil())
				Expect(session.ID).To(Equal(sessionID))

				hashKey := fmt.Sprintf("level:sessions:%s", sessionID)
				exists, err := testClient.Exists(hashKey).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())

				lastUpdated, err := testClient.HGet(hashKey, sessionManager.GetLastUpdatedKey()).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(lastUpdated).NotTo(BeNil())

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Starting new session...",
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
				_, err := sm.Start(sessionID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("connection refused"))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Starting new session...",
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

		Describe("Merging sessions", func() {
			It("should merge a session into another one", func() {
				sm := getDefaultSM(logger)

				oldSessionID := uuid.NewV4().String()
				sm.Start(oldSessionID)
				hashKey := fmt.Sprintf("level:sessions:%s", oldSessionID)
				testClient.HSet(hashKey, "someKey", "someValue")

				sessionID := uuid.NewV4().String()
				sm.Start(sessionID)
				count, err := sm.Merge(oldSessionID, sessionID)
				Expect(err).NotTo(HaveOccurred())
				Expect(count).To(Equal(1))

				exists, err := testClient.Exists(hashKey).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeFalse())

				hashKey = fmt.Sprintf("level:sessions:%s", sessionID)
				exists, err = testClient.Exists(hashKey).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())

				someValue, err := testClient.HGet(hashKey, "someKey").Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(someValue).To(Equal("someValue"))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Merging sessions...",
					"source", "sessionManager",
					"operation", "Merge",
					"oldSessionID", oldSessionID,
					"sessionID", sessionID,
				))

				Expect(logger).To(HaveLogMessage(
					zap.InfoLevel, "Sessions merged successfully.",
					"source", "sessionManager",
					"operation", "Merge",
					"oldSessionID", oldSessionID,
					"sessionID", sessionID,
				))
			})

			It("should not merge a non-existing session", func() {
				sm := getDefaultSM(logger)

				sessionID := uuid.NewV4().String()
				sm.Start(sessionID)

				count, err := sm.Merge("invalid-id", sessionID)
				Expect(count).To(Equal(0))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Session with session ID invalid-id was not found in session storage."))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Merging sessions...",
					"source", "sessionManager",
					"operation", "Merge",
				))

				Expect(logger).To(HaveLogMessage(
					zap.ErrorLevel, "Previous session was not found.",
					"source", "sessionManager",
					"operation", "Merge",
					"error", "Session was not found!",
				))
			})

			It("should not merge if invalid connection", func() {
				sm := getDefaultSM(logger)

				sessionID := uuid.NewV4().String()
				sm.Start(sessionID)

				sm.Client = getFaultyRedisClient()

				count, err := sm.Merge("invalid-id", sessionID)
				Expect(count).To(Equal(0))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("connection refused"))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Merging sessions...",
					"source", "sessionManager",
					"operation", "Merge",
				))

				Expect(logger).To(HaveLogMessage(
					zap.ErrorLevel, "Merging sessions failed.",
					"source", "sessionManager",
					"operation", "Merge",
					"error", "dial tcp 0.0.0.0:9876: getsockopt: connection refused",
				))
			})

		})

		Describe("Loading sessions", func() {
			It("should be able to load a session", func() {
				sm := getDefaultSM(logger)
				sessionID := uuid.NewV4().String()
				sm.Start(sessionID)

				session, err := sm.Load(sessionID)

				Expect(err).NotTo(HaveOccurred())

				Expect(session.ID).To(Equal(sessionID))
				Expect(session.Manager).To(BeEquivalentTo(sm))
				Expect(session.Get(sessionManager.GetLastUpdatedKey())).To(BeNumerically(">", 0))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Loading session...",
					"source", "sessionManager",
					"operation", "Load",
					"sessionID", sessionID,
				))

				Expect(logger).To(HaveLogMessage(
					zap.InfoLevel, "Session loaded successfully.",
					"source", "sessionManager",
					"operation", "Load",
					"sessionID", sessionID,
				))
			})

			It("should not load a session if invalid id", func() {
				sm := getDefaultSM(logger)
				sessionID := uuid.NewV4().String()

				session, err := sm.Load(sessionID)
				Expect(session).To(BeNil())
				Expect(err).To(HaveOccurred())
				expected := fmt.Sprintf("Session with session ID %s was not found in session storage.", sessionID)
				Expect(err.Error()).To(Equal(expected))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Loading session...",
					"source", "sessionManager",
					"operation", "Load",
					"sessionID", sessionID,
				))
			})
			It("should not load a session if invalid connection", func() {
				sm := getDefaultSM(logger)
				sessionID := uuid.NewV4().String()
				sm.Start(sessionID)

				sm.Client = getFaultyRedisClient()
				session, err := sm.Load(sessionID)
				Expect(session).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("connection refused"))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Loading session...",
					"source", "sessionManager",
					"operation", "Load",
					"sessionID", sessionID,
				))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Reloading session...",
					"source", "sessionManager",
					"operation", "ReloadSession",
					"sessionID", sessionID,
				))

				Expect(logger).To(HaveLogMessage(
					zap.ErrorLevel, "Reloading session failed.",
					"source", "sessionManager",
					"operation", "ReloadSession",
					"lastUpdatedKey", "__last_updated__",
					"sessionKey", fmt.Sprintf("level:sessions:%s", sessionID),
					"error", "dial tcp 0.0.0.0:9876: getsockopt: connection refused",
				))
			})
		})

		Describe("Reloading sessions", func() {
			It("should be able to reload a session", func() {
				sm := getDefaultSM(logger)
				sessionID := uuid.NewV4().String()
				sm.Start(sessionID)

				session := &sessionManager.Session{
					ID:      sessionID,
					Manager: sm,
				}
				err := sm.ReloadSession(session)
				Expect(err).NotTo(HaveOccurred())

				Expect(session.ID).To(Equal(sessionID))
				Expect(session.Get(sessionManager.GetLastUpdatedKey())).To(BeNumerically(">", 0))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Reloading session...",
					"source", "sessionManager",
					"operation", "ReloadSession",
					"sessionID", sessionID,
				))

				Expect(logger).To(HaveLogMessage(
					zap.InfoLevel, "Session reloaded successfully.",
					"source", "sessionManager",
					"operation", "ReloadSession",
					"sessionID", sessionID,
				))
			})

			It("should not reload a session if invalid id", func() {
				sm := getDefaultSM(logger)
				sessionID := uuid.NewV4().String()

				session := &sessionManager.Session{
					ID:      sessionID,
					Manager: sm,
				}
				err := sm.ReloadSession(session)
				Expect(err).To(HaveOccurred())
				expected := fmt.Sprintf("Session with session ID %s was not found in session storage.", sessionID)
				Expect(err.Error()).To(Equal(expected))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Reloading session...",
					"source", "sessionManager",
					"operation", "ReloadSession",
					"sessionID", sessionID,
				))

				Expect(logger).To(HaveLogMessage(
					zap.ErrorLevel, "Session was not found!",
					"source", "sessionManager",
					"operation", "ReloadSession",
					"lastUpdatedKey", "__last_updated__",
					"sessionKey", fmt.Sprintf("level:sessions:%s", sessionID),
					"error", expected,
				))
			})

			It("should not reload a session if invalid connection", func() {
				sm := getDefaultSM(logger)
				sessionID := uuid.NewV4().String()

				session := &sessionManager.Session{
					ID:      sessionID,
					Manager: sm,
				}
				sm.Client = getFaultyRedisClient()
				err := sm.ReloadSession(session)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("connection refused"))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Reloading session...",
					"source", "sessionManager",
					"operation", "ReloadSession",
					"sessionID", sessionID,
				))

				Expect(logger).To(HaveLogMessage(
					zap.ErrorLevel, "Reloading session failed.",
					"source", "sessionManager",
					"operation", "ReloadSession",
					"lastUpdatedKey", "__last_updated__",
					"sessionKey", fmt.Sprintf("level:sessions:%s", sessionID),
					"error", "dial tcp 0.0.0.0:9876: getsockopt: connection refused",
				))
			})
		})

		Describe("Validating sessions", func() {
			It("should validate a new session", func() {
				sm := getDefaultSM(logger)
				sessionID := uuid.NewV4().String()
				_, err := sm.Start(sessionID)
				Expect(err).NotTo(HaveOccurred())

				session, err := sm.Load(sessionID)
				Expect(err).NotTo(HaveOccurred())

				valid, err := sm.ValidateSession(session)
				Expect(err).NotTo(HaveOccurred())
				Expect(valid).To(BeTrue())

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Validating session...",
					"source", "sessionManager",
					"operation", "ValidateSession",
					"sessionID", sessionID,
				))

				Expect(logger).To(HaveLogMessage(
					zap.InfoLevel, "Session validated successfully.",
					"source", "sessionManager",
					"operation", "ValidateSession",
					"sessionID", sessionID,
					"isValid", true,
				))
			})

			It("should invalidate an old session", func() {
				sm := getDefaultSM(logger)
				sessionID := uuid.NewV4().String()
				_, err := sm.Start(sessionID)
				Expect(err).NotTo(HaveOccurred())

				session := &sessionManager.Session{
					ID:          sessionID,
					Manager:     sm,
					LastUpdated: 0,
				}
				valid, err := sm.ValidateSession(session)
				Expect(err).NotTo(HaveOccurred())
				Expect(valid).To(BeFalse())

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Validating session...",
					"source", "sessionManager",
					"operation", "ValidateSession",
					"sessionID", sessionID,
				))

				Expect(logger).To(HaveLogMessage(
					zap.InfoLevel, "Session validated successfully.",
					"source", "sessionManager",
					"operation", "ValidateSession",
					"sessionID", sessionID,
					"isValid", false,
				))
			})

			It("should fail if bad connection", func() {
				sm := getDefaultSM(logger)
				sessionID := uuid.NewV4().String()
				_, err := sm.Start(sessionID)
				Expect(err).NotTo(HaveOccurred())

				session := &sessionManager.Session{
					ID:      sessionID,
					Manager: sm,
				}

				sm.Client = getFaultyRedisClient()

				_, err = sm.ValidateSession(session)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("connection refused"))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Validating session...",
					"source", "sessionManager",
					"operation", "ValidateSession",
					"sessionID", sessionID,
				))

				Expect(logger).To(HaveLogMessage(
					zap.ErrorLevel, "Could not validate session.",
					"source", "sessionManager",
					"operation", "ValidateSession",
					"sessionID", sessionID,
					"error", "dial tcp 0.0.0.0:9876: getsockopt: connection refused",
				))
			})

			It("should fail if corrupt data", func() {
				sm := getDefaultSM(logger)
				sessionID := uuid.NewV4().String()
				_, err := sm.Start(sessionID)
				Expect(err).NotTo(HaveOccurred())

				session := &sessionManager.Session{
					ID:      sessionID,
					Manager: sm,
				}

				hashKey := fmt.Sprintf("level:sessions:%s", sessionID)
				_, err = testClient.HSet(hashKey, sessionManager.GetLastUpdatedKey(), "qwe").Result()
				Expect(err).NotTo(HaveOccurred())

				_, err = sm.ValidateSession(session)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("parsing \"qwe\": invalid syntax"))

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Validating session...",
					"source", "sessionManager",
					"operation", "ValidateSession",
					"sessionID", sessionID,
				))

				Expect(logger).To(HaveLogMessage(
					zap.ErrorLevel, "Could not validate session (invalid timestamp).",
					"source", "sessionManager",
					"operation", "ValidateSession",
					"sessionID", sessionID,
					"error", "strconv.ParseInt: parsing \"qwe\": invalid syntax",
				))
			})
		})
	})
})
