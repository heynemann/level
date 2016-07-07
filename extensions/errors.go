// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package extensions

import "fmt"

//SessionNotFoundError identified an error that occurred because a given session was not found.
type SessionNotFoundError struct {
	SessionID string
}

func (s *SessionNotFoundError) Error() string {
	return fmt.Sprintf("Session with session ID %s was not found in session storage.", s.SessionID)
}

//UnserializableItemError indicates that an item that could not be serialized was added to a session
type UnserializableItemError struct {
	SessionID string
	Item      interface{}
}

func (u *UnserializableItemError) Error() string {
	return fmt.Sprintf("Could not serialize/deserialize value %v for session with session ID %s.", u.Item, u.SessionID)
}
