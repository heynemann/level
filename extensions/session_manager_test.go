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
	"gopkg.in/vmihailenco/msgpack.v2"

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

var _ = Describe("Session Management", func() {

	var testClient *redis.Client

	BeforeEach(func() {
		testClient = redis.NewClient(&redis.Options{
			Addr:     "localhost:7777",
			Password: "",
			DB:       0,
		})
	})

	Describe("Session Manager", func() {

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

				Expect(err).To(HaveOccurred())
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
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Session with session ID invalid-id was not found in session storage."))
			})
		})

		Describe("can get session", func() {
			It("should be able to load a session", func() {
				sm := getDefaultSM()
				sessionID := uuid.NewV4().String()
				sm.Start(sessionID)

				session, err := sm.Load(sessionID)

				Expect(err).ToNot(HaveOccurred())

				Expect(session.ID).To(Equal(sessionID))
				Expect(session.Manager).To(BeEquivalentTo(sm))
				Expect(session.Get("lastupdated")).To(BeNumerically(">", 0))
			})
		})
	})

	Describe("Session", func() {
		Describe("can serialize", func() {
			It("should serialize using msgpack", func() {
				expected := map[string]interface{}{"a": 1}
				s := extensions.Session{}
				result, err := s.Serialize(expected)

				Expect(err).NotTo(HaveOccurred())

				serialized, err := msgpack.Marshal(expected)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeEquivalentTo(serialized))
			})

			It("should fail to serialize invalid object", func() {
				s := extensions.Session{}
				_, err := s.Serialize(func() {})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("msgpack: Encode(unsupported func())"))
			})
		})

		Describe("can deserialize", func() {
			It("should deserialize using msgpack", func() {
				expected := map[interface{}]interface{}{"a": 1}
				serialized, err := msgpack.Marshal(expected)
				Expect(err).NotTo(HaveOccurred())

				s := extensions.Session{}
				result, err := s.Deserialize(string(serialized))
				actual := result.(map[interface{}]interface{})

				Expect(err).NotTo(HaveOccurred())
				Expect(actual["a"]).To(BeEquivalentTo(1))
			})

			It("should fail to deserialize invalid payload", func() {
				s := extensions.Session{}
				_, err := s.Deserialize("")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("EOF"))
			})
		})
	})
})
