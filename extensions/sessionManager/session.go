// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package sessionManager

import "fmt"

//Session represents an user session
type Session struct {
	ID          string
	Manager     SessionManager
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

//Get returns an item in the session
func (session *Session) Get(key string) interface{} {
	valid, err := session.Manager.ValidateSession(session)
	if err != nil {
		return nil
	}

	if !valid {
		session.Manager.ReloadSession(session)
	}
	return session.data[key]
}

//Set stores a value in session and updates the timestamp for the session object
func (session *Session) Set(key string, value interface{}) error {
	err := session.Manager.SetKey(session, key, value)
	if err != nil {
		return err
	}
	return nil
}
