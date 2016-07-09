// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package sessionManager

import (
	"strconv"
	"time"

	"github.com/heynemann/level/extensions"
	"github.com/heynemann/level/extensions/redis"
	"github.com/uber-go/zap"

	redisCli "gopkg.in/redis.v4"
)

// SessionManager is responsible for handling session data
type SessionManager struct {
	Expiration int
	Logger     zap.Logger
	Client     *redisCli.Client
}

//GetSessionManager returns a connected SessionManager ready to be used.
func GetSessionManager(redisHost string, redisPort int, redisPass string, redisDB int, expiration int, logger zap.Logger) (*SessionManager, error) {
	l := logger.With(
		zap.String("source", "sessionManager"),
		zap.Duration("expiration", time.Duration(expiration)*time.Second),
	)

	sessionManager := &SessionManager{
		Expiration: expiration,
		Logger:     l,
	}

	cli, err := redis.New(redisHost, redisPort, redisPass, redisDB, l)
	if err != nil {
		return nil, err
	}
	sessionManager.Client = cli

	return sessionManager, nil
}

//Start starts a new session in the storage (or resumes an old one)
func (s *SessionManager) Start(sessionID string) error {
	l := s.Logger.With(zap.String("operation", "Start"))
	hashKey := getSessionKey(sessionID)
	timestamp := time.Now().UnixNano()

	l.Debug("Starting new session.", zap.String("sessionID", hashKey), zap.Int64("sessionTimestamp", timestamp))
	script := `
		local res
		res = redis.call("HSET", KEYS[1], KEYS[2], ARGV[1])
		res = redis.call("EXPIRE", KEYS[1], ARGV[2])
		return res
	`
	startSessionScript := redisCli.NewScript(script)
	expire := strconv.FormatInt(int64(s.Expiration), 10)
	start := time.Now()
	_, err := startSessionScript.Run(
		s.Client, []string{hashKey, GetLastUpdatedKey()},
		strconv.FormatInt(timestamp, 10), expire,
	).Result()
	if err != nil {
		l.Error("Could not start session.", zap.Error(err))
		return err
	}
	l.Info("Started session successfully.", zap.Duration("sessionStartDuration", time.Now().Sub(start)))

	return nil
}

//Merge gets all the keys from old session into new session (no overwrites done).
func (s *SessionManager) Merge(oldSessionID, sessionID string) (int, error) {
	l := s.Logger.With(
		zap.String("operation", "Merge"),
		zap.String("oldSessionID", oldSessionID),
		zap.String("sessionID", sessionID),
	)

	oldHashKey := getSessionKey(oldSessionID)
	hashKey := getSessionKey(sessionID)

	l.Debug("Merging sessions.")
	mergeScript := redisCli.NewScript(`
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
	start := time.Now()
	totalKeys, err := mergeScript.Run(s.Client, []string{oldHashKey, hashKey}).Result()
	if err != nil {
		if err.Error() == "Session was not found!" {
			l.Error("Previous session was not found.", zap.Error(err))
			return 0, &extensions.SessionNotFoundError{SessionID: oldSessionID}
		}
		l.Error("Merging sessions failed.", zap.Error(err))
		return 0, err
	}
	l.Info("Sessions merged successfully.", zap.Duration("sessionMergeDuration", time.Now().Sub(start)))
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
