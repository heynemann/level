// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package heartbeat_test

import (
	"strconv"
	"time"

	"gopkg.in/redis.v4"

	"github.com/heynemann/level/extensions/heartbeat"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Heartbeat", func() {
	var testClient *redis.Client

	BeforeEach(func() {
		testClient = redis.NewClient(&redis.Options{
			Addr:     "localhost:7777",
			Password: "",
			DB:       0,
		})
	})

	Describe("Heartbeat Extension", func() {
		Describe("Heartbeat creation", func() {
			It("should create new heartbeat", func() {
				heartbeat, err := heartbeat.New("some-server", "localhost", 7777, "", 0, 10*time.Second)
				Expect(err).NotTo(HaveOccurred())
				Expect(heartbeat).NotTo(BeNil())
			})
		})

		Describe("Heartbeat registry", func() {
			It("should register server in redis", func() {
				heartbeat, err := heartbeat.New("some-server", "localhost", 7777, "", 0, 10*time.Second)
				Expect(err).NotTo(HaveOccurred())

				err = heartbeat.Register()
				Expect(err).NotTo(HaveOccurred())

				serverKey := "server-status:some-server"
				result, err := testClient.Get(serverKey).Result()
				Expect(err).NotTo(HaveOccurred())
				value, err := strconv.ParseInt(result, 10, 64)
				Expect(err).NotTo(HaveOccurred())
				Expect(int(value)).To(BeNumerically(">", 0))

				list, err := testClient.SMembers("available-servers").Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(list[0]).To(Equal("some-server"))
			})
		})
	})
})
