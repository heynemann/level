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
	Type      string
	Timestamp time.Time
	Payload   map[string]interface{}
}

//NewEvent builds an event and returns it
func NewEvent(eventType string, payload map[string]interface{}) *Event {
	return &Event{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}
