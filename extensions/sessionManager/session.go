// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package sessionManager

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/heynemann/level/extensions"
)

//Session represents an user session
type Session struct {
	ID          string
	Manager     *SessionManager
	LastUpdated int64
	data        map[string]interface{}
}

func getSessionKey(sessionID string) string {
	return fmt.Sprintf("level:sessions:%s", sessionID)
}

//GetLastUpdatedKey returns the key to use for last updated timestamp in sessions
func GetLastUpdatedKey() string {
	return "__last_updated__"
}

func (session *Session) validateTimestamp() bool {
	lastUpdatedKey := GetLastUpdatedKey()
	hashKey := getSessionKey(session.ID)
	ts, err := session.Manager.Client.HGet(hashKey, lastUpdatedKey).Result()
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
		session.Manager.ReloadSession(session)
	}
	return session.data[key]
}

//Set stores a value in session and updates the timestamp for the session object
func (session *Session) Set(key string, value interface{}) error {
	lastUpdatedKey := GetLastUpdatedKey()
	hashKey := getSessionKey(session.ID)
	serialized, err := extensions.Serialize(value)
	if err != nil {
		if strings.HasPrefix(err.Error(), "msgpack: Encode(unsupported") {
			return &extensions.UnserializableItemError{SessionID: session.ID, Item: value}
		}
		return err
	}

	ts := time.Now().UnixNano()

	_, err = session.Manager.Client.HMSet(hashKey, map[string]string{
		key:            serialized,
		lastUpdatedKey: strconv.FormatInt(ts, 10),
	}).Result()

	if err != nil {
		return err
	}

	session.data[key] = value
	session.LastUpdated = ts

	return nil
}
