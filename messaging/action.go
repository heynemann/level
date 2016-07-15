// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package messaging

import "time"

//Action represents an action from the client to the server
type Action struct {
	SessionID string
	Type      string
	Key       string
	Timestamp time.Time
	Payload   map[string]interface{}
}

//NewAction build an action and returns it
func NewAction(sessionID string, key string, payload map[string]interface{}) *Action {
	return &Action{
		Type:      "action",
		Key:       key,
		SessionID: sessionID,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}
