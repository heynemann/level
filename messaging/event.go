// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package messaging

import "time"

//Event represents an event from the server to the clients
type Event struct {
	Key       string
	Timestamp time.Time
	Payload   interface{}
}

//NewEvent builds an event and returns it
func NewEvent(key string, payload interface{}) *Event {
	return &Event{
		Key:       key,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}
