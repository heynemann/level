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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/uber-go/zap"
)

var _ = Describe("Channel", func() {
	var logger *MockLogger

	BeforeEach(func() {
		logger = NewMockLogger()
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

				Expect(channel.Config.GetString("channel.workingString")).To(Equal("WORKING"))
			})
		})
	})
})
