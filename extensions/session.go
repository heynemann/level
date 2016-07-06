// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package extensions

import (
	"fmt"
	"strconv"
)

//Session represents an user session
type Session struct {
	ID          string
	Manager     *SessionManager
	LastUpdated int64
	data        map[string]interface{}
}

func getSessionKey(sessionID string) string {
	return fmt.Sprintf("session-%s", sessionID)
}

//Reload reloads the data in session
func (session *Session) Reload() error {
	all, err := session.Manager.client.HGetAll(getSessionKey(session.ID)).Result()
	if err != nil {
		return err
	}
	if len(all) == 0 {
		return &SessionNotFoundError{session.ID}
	}

	session.data = map[string]interface{}{}

	for k, v := range all {
		if k == "lastupdated" {
			if lastUpdated, err := strconv.ParseInt(v, 10, 64); err == nil {
				session.LastUpdated = lastUpdated
			}
		}
		item, err := Deserialize(v)
		if err != nil {
			continue
		}
		session.data[k] = item
	}

	return nil
}

func (session *Session) validateTimestamp() bool {
	hashKey := getSessionKey(session.ID)
	ts, err := session.Manager.client.HGet(hashKey, "lastupdated").Result()
	if err != nil {
		return false
	}

	timestamp, err := strconv.ParseInt(ts, 10, 64)
	if err != nil || timestamp != session.LastUpdated {
		return false
	}

	return true
}

//Get returns an item in the session
func (session *Session) Get(key string) interface{} {
	if !session.validateTimestamp() {
		session.Reload()
	}
	return session.data[key]
}
