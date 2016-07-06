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
	sessionID string
}

func (s *SessionNotFoundError) Error() string {
	return fmt.Sprintf("Session with session ID %s was not found in session storage.", s.sessionID)
}
