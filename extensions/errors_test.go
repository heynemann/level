// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package extensions_test

import (
	"github.com/heynemann/level/extensions"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Custom Errors", func() {

	Describe("Session Not Found Error", func() {
		It("Should have proper error message", func() {
			sessionID := "something"

			err := &extensions.SessionNotFoundError{
				SessionID: sessionID,
			}

			Expect(err.Error()).To(Equal("Session with session ID something was not found in session storage."))
		})
	})

	Describe("Unserializable Item Error", func() {
		It("Should have proper error message", func() {
			sessionID := "something"
			item := "qwe"

			err := &extensions.UnserializableItemError{
				SessionID: sessionID,
				Item:      item,
			}

			Expect(err.Error()).To(Equal("Could not serialize/deserialize value qwe for session with session ID something."))
		})
	})

})
