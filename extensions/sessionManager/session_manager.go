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
	"time"

	"github.com/heynemann/level/extensions"

	"gopkg.in/redis.v4"
)

// SessionManager is responsible for handling session data
type SessionManager struct {
	Expiration int
	client     *redis.Client
}

//GetSessionManager returns a connected SessionManager ready to be used.
func GetSessionManager(redisHost string, redisPort int, redisPass string, redisDB int, expiration int) (*SessionManager, error) {
	sessionManager := &SessionManager{
		Expiration: expiration,
	}

	sessionManager.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
		Password: redisPass,
		DB:       redisDB,
	})

	_, err := sessionManager.client.Ping().Result()
	if err != nil {
		return nil, err
	}

	return sessionManager, nil
}

//Start starts a new session in the storage (or resumes an old one)
func (s *SessionManager) Start(sessionID string) error {
	hashKey := getSessionKey(sessionID)
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	mergeScript := redis.NewScript(`
		redis.call("HSET", KEYS[1], KEYS[2], ARGV[1])
		redis.call("EXPIRE", KEYS[1], ARGV[2])
		return null
	`)
	expire := strconv.FormatInt(int64(s.Expiration), 10)
	_, err := mergeScript.Run(s.client, []string{hashKey, GetLastUpdatedKey()}, timestamp, expire).Result()
	if err != nil {
		return err
	}

	return nil
}

//Merge gets all the keys from old session into new session (no overwrites done).
func (s *SessionManager) Merge(oldSessionID, sessionID string) (int, error) {
	oldHashKey := getSessionKey(oldSessionID)
	hashKey := getSessionKey(sessionID)

	mergeScript := redis.NewScript(`
		local values = redis.call("HGETALL", KEYS[1])
		if (#values == 0) then
			return redis.error_reply("Session was not found!")
		end
		redis.call("DEL", KEYS[1])

		local keys = 0
		local res
		for i=1, #values, 2 do
			res = redis.call("HSETNX", KEYS[2], values[i], values[i + 1])
			keys = keys + res
		end

		return keys
	`)
	totalKeys, err := mergeScript.Run(s.client, []string{oldHashKey, hashKey}).Result()
	if err != nil {
		if err.Error() == "Session was not found!" {
			return 0, &extensions.SessionNotFoundError{SessionID: oldSessionID}
		}
		return 0, err
	}
	return int(totalKeys.(int64)), nil
}

//Load loads a session from the storage with all its items
func (s *SessionManager) Load(sessionID string) (*Session, error) {
	sess := &Session{
		ID:      sessionID,
		Manager: s,
		data:    make(map[string]interface{}),
	}
	err := sess.Reload()
	if err != nil {
		return nil, err
	}

	return sess, nil
}
