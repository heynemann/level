// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package channel_test

import (
	"github.com/heynemann/level/channel"
	. "github.com/heynemann/level/testing"
	gnatsdServer "github.com/nats-io/gnatsd/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/uber-go/zap"
)

var _ = Describe("Channel", func() {
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

	Describe("Channel", func() {
		Describe("Channel creation", func() {
			It("should create new channel", func() {
				channel, err := channel.New(nil, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(channel).NotTo(BeNil())
				Expect(channel.ServerOptions).NotTo(BeNil())
				Expect(channel.ServerOptions.Host).To(Equal("0.0.0.0"))
				Expect(channel.ServerOptions.Port).To(Equal(3000))
				Expect(channel.ServerOptions.Debug).To(BeTrue())

				Expect(channel.Config).NotTo(BeNil())

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Initializing channel...",
					"source", "channel",
					"host", "0.0.0.0",
					"port", 3000,
					"operation", "initializeChannel",
				))

				Expect(logger).To(HaveLogMessage(
					zap.InfoLevel, "Channel initialized successfully.",
					"source", "channel",
					"host", "0.0.0.0",
					"port", 3000,
					"operation", "initializeChannel",
				))
			})
		})

		Describe("Channel Default Configurations", func() {
			It("Should set default configurations", func() {
				channel, err := channel.New(nil, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(channel).NotTo(BeNil())

				Expect(channel.Config.GetString("channel.workingText")).To(Equal("WORKING"))
			})
		})

		Describe("Channel Load Configuration", func() {
			It("Should load configuration from file", func() {
				options := channel.DefaultOptions()
				options.ConfigFile = "../config/default.yaml"

				channel, err := channel.New(options, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(channel.Config).NotTo(BeNil())

				expected := channel.Config.GetString("channel.services.redis.host")
				Expect(expected).To(Equal("127.0.0.1"))
			})

			It("Should fail with non-existent config file", func() {
				options := channel.DefaultOptions()
				options.ConfigFile = "../config/does-not-exist.yaml"

				channel, err := channel.New(options, logger)
				Expect(channel).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no such file or directory"))
			})

			It("Should fail with invalid yaml", func() {
				options := channel.DefaultOptions()
				options.ConfigFile = "../config/invalid.yaml"

				channel, err := channel.New(options, logger)
				Expect(channel).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("yaml: unmarshal errors"))
			})
		})

		Describe("Channel Initialization", func() {
			It("should initialize redis", func() {
				channel, err := channel.New(nil, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(channel).NotTo(BeNil())

				Expect(channel.Redis).NotTo(BeNil())

				Expect(logger).To(HaveLogMessage(
					zap.DebugLevel, "Connecting to Redis...",
					"source", "channel",
				))

				Expect(logger).To(HaveLogMessage(
					zap.InfoLevel, "Connected to Redis successfully.",
					"source", "channel",
				))
			})

			It("should fail if invalid connection to redis", func() {
				options := channel.NewOptions(
					"0.0.0.0",
					3000,
					true,
					"../config/invalid-redis.yaml",
				)
				channel, err := channel.New(options, logger)
				Expect(channel).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("connection refused"))
			})

			It("should initialize pubsub", func() {
				channel, err := channel.New(nil, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(channel).NotTo(BeNil())

				Expect(channel.PubSub).NotTo(BeNil())
			})

			It("should initialize pubsub with invalid nsq", func() {
				options := channel.NewOptions(
					"0.0.0.0",
					3000,
					true,
					"../config/invalid-nats.yaml",
				)
				channel, err := channel.New(options, logger)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("nats: no servers available for connection"))
				Expect(channel).To(BeNil())
			})

		})
	})
})
