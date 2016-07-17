// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package session_test

import (
	"time"

	"github.com/heynemann/level/extensions/pubsub"
	"github.com/heynemann/level/messaging"
	"github.com/heynemann/level/services/session"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
)

var _ = Describe("Session Service", func() {
	It("Should handle session action", func() {
		service := session.NewSessionService()
		ps := &pubsub.PubSub{}
		service.Initialize(ps)
		Expect(service.PubSub).To(Equal(ps))

		action := messaging.NewAction(
			"",
			"channel.session.start",
			nil,
		)

		Expect(action).NotTo(BeNil())

		event := &messaging.Event{}
		reply := func(ev *messaging.Event) error {
			event = ev
			return nil
		}

		sessionID := uuid.NewV4().String()
		service.HandleAction(sessionID, action, reply, time.Now().UnixNano())
		Expect(event.Key).To(Equal("channel.session.joined"))
		Expect(event.Payload).NotTo(BeNil())

		p := event.Payload.(map[string]interface{})
		Expect(p["sessionID"]).To(Equal(sessionID))
	})
})
