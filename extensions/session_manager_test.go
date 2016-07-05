// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package extensions_test

import (
	"fmt"

	"gopkg.in/redis.v4"

	"github.com/heynemann/level/extensions"
	"github.com/satori/go.uuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getDefaultSM() *extensions.SessionManager {
	sessionManager, _ := extensions.GetSessionManager(
		"localhost", // Redis Host
		7777,        // Redis Port
		"",          // Redis Pass
		0,           // Redis DB
	)

	return sessionManager
}

var _ = Describe("Session Manager", func() {

	var testClient *redis.Client

	BeforeEach(func() {
		testClient = redis.NewClient(&redis.Options{
			Addr:     "localhost:7777",
			Password: "",
			DB:       0,
		})
	})

	Describe("Initialized properly", func() {

		It("When getting a new instance", func() {
			sessionManager, err := extensions.GetSessionManager(
				"localhost", // Redis Host
				7777,        // Redis Port
				"",          // Redis Pass
				0,           // Redis DB
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(sessionManager).NotTo(BeNil())
		})

		It("should be connected to Redis", func() {
			_, err := extensions.GetSessionManager(
				"localhost", // Redis Host
				7777,        // Redis Port
				"",          // Redis Pass
				0,           // Redis DB
			)

			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Initialized with wrong params", func() {
		It("should not be connected to Redis", func() {
			sessionManager, err := extensions.GetSessionManager(
				"localhost", // Redis Host
				1249,        // Redis Port
				"",          // Redis Pass
				0,           // Redis DB
			)

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
			Expect(sessionManager).To(BeNil())
		})
	})

	Describe("can start sessions", func() {
		It("should start a session when provided with session id", func() {
			sm := getDefaultSM()
			sessionID := uuid.NewV4().String()
			sm.Start(sessionID)

			hashKey := fmt.Sprintf("session-%s", sessionID)
			exists, err := testClient.Exists(hashKey).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())

			lastUpdated, err := testClient.HGet(hashKey, "lastupdated").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(lastUpdated).NotTo(BeNil())
		})
	})

	Describe("can merge sessions", func() {
		It("should merge a session into another one", func() {
			sm := getDefaultSM()

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
			sm := getDefaultSM()

			sessionID := uuid.NewV4().String()
			sm.Start(sessionID)

			count, err := sm.Merge("invalid-id", sessionID)
			Expect(count).To(Equal(0))
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("Session with session ID invalid-id was not found in session storage."))
		})
	})

	Describe("can set new keys", func() {
		It("should be able to set a new key", func() {

		})
	})
})
