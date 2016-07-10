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
)

var _ = Describe("Channel", func() {
	var logger *MockLogger

	BeforeEach(func() {
		logger = NewMockLogger()
	})

	Describe("Channel", func() {
		Describe("Channel creation", func() {
			It("should create new channel", func() {
				channel, err := channel.New(logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(channel).NotTo(BeNil())
			})
		})
	})
})
