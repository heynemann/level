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

var _ = Describe("Session Management", func() {

	var testClient *redis.Client

	BeforeEach(func() {
		testClient = redis.NewClient(&redis.Options{
			Addr:     "localhost:7777",
			Password: "",
			DB:       0,
		})
	})

	Describe("Session", func() {
		Describe("can serialize", func() {
			It("should serialize using msgpack", func() {
				expected := map[string]interface{}{"a": 1}
				result, err := extensions.Serialize(expected)

				Expect(err).NotTo(HaveOccurred())

				serialized, err := msgpack.Marshal(expected)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(BeEquivalentTo(serialized))
			})

			It("should fail to serialize invalid object", func() {
				_, err := extensions.Serialize(func() {})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("msgpack: Encode(unsupported func())"))
			})
		})

		Describe("can deserialize", func() {
			It("should deserialize using msgpack", func() {
				expected := map[interface{}]interface{}{"a": 1}
				serialized, err := msgpack.Marshal(expected)
				Expect(err).NotTo(HaveOccurred())

				result, err := extensions.Deserialize(string(serialized))
				actual := result.(map[interface{}]interface{})

				Expect(err).NotTo(HaveOccurred())
				Expect(actual["a"]).To(BeEquivalentTo(1))
			})

			It("should fail to deserialize invalid payload", func() {
				_, err := extensions.Deserialize("")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("EOF"))
			})
		})

		Describe("can get data", func() {
			It("should get items in session", func() {
				sm := getDefaultSM()
				sessionID := uuid.NewV4().String()
				sm.Start(sessionID)
				sessionKey := fmt.Sprintf("session-%s", sessionID)

				serialized, err := extensions.Serialize("someValue")
				Expect(err).NotTo(HaveOccurred())
				testClient.HMSet(sessionKey, map[string]string{"someKey": serialized}).Result()

				session, err := sm.Load(sessionID)
				Expect(err).NotTo(HaveOccurred())

				value := session.Get("someKey")
				Expect(value).To(Equal("someValue"))
			})

			It("should reload session when getting items", func() {
				sm := getDefaultSM()
				sessionID := uuid.NewV4().String()
				sm.Start(sessionID)
				sessionKey := fmt.Sprintf("session-%s", sessionID)

				serialized, err := extensions.Serialize("someValue")
				Expect(err).NotTo(HaveOccurred())
				testClient.HMSet(sessionKey, map[string]string{"someKey": serialized}).Result()

				session, err := sm.Load(sessionID)
				Expect(err).NotTo(HaveOccurred())

				serialized, err = extensions.Serialize("otherValue")
				Expect(err).NotTo(HaveOccurred())
				_, err = testClient.HSet(sessionKey, "lastupdated", "2").Result()
				Expect(err).NotTo(HaveOccurred())
				_, err = testClient.HSet(sessionKey, "someKey", serialized).Result()
				Expect(err).NotTo(HaveOccurred())

				value := session.Get("someKey")
				Expect(value).To(Equal("otherValue"))
			})
		})
	})
})
