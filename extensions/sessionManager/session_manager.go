// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package sessionManager

import (
	"strconv"
	"strings"
	"time"

	"github.com/heynemann/level/extensions"
	"github.com/heynemann/level/extensions/redis"
	"github.com/uber-go/zap"

	redisCli "gopkg.in/redis.v4"
)

type SessionManager interface {
	Start(sessionID string) error
	Merge(oldSessionID, sessionID string) (int, error)
	Load(sessionID string) (*Session, error)
	ReloadSession(session *Session) error
	ValidateSession(session *Session) (bool, error)
	SetKey(session *Session, key string, value interface{}) error
}

// SessionManager is responsible for handling session data
type RedisSessionManager struct {
	Expiration int
	Logger     zap.Logger
	Client     *redisCli.Client
}

//GetSessionManager returns a connected SessionManager ready to be used.
func GetRedisSessionManager(redisHost string, redisPort int, redisPass string, redisDB int, expiration int, logger zap.Logger) (*RedisSessionManager, error) {
	l := logger.With(
		zap.String("source", "sessionManager"),
		zap.Duration("expiration", time.Duration(expiration)*time.Second),
	)

	sessionManager := &RedisSessionManager{
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
func (s *RedisSessionManager) Start(sessionID string) error {
	l := s.Logger.With(zap.String("operation", "Start"))
	hashKey := getSessionKey(sessionID)
	timestamp := time.Now().UnixNano()

	l.Debug("Starting new session...", zap.String("sessionID", hashKey), zap.Int64("sessionTimestamp", timestamp))
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
func (s *RedisSessionManager) Merge(oldSessionID, sessionID string) (int, error) {
	l := s.Logger.With(
		zap.String("operation", "Merge"),
		zap.String("oldSessionID", oldSessionID),
		zap.String("sessionID", sessionID),
	)

	oldHashKey := getSessionKey(oldSessionID)
	hashKey := getSessionKey(sessionID)

	l.Debug("Merging sessions...")
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
func (s *RedisSessionManager) Load(sessionID string) (*Session, error) {
	l := s.Logger.With(
		zap.String("operation", "Load"),
		zap.String("sessionID", sessionID),
	)

	sess := &Session{
		ID:      sessionID,
		Manager: s,
		data:    make(map[string]interface{}),
	}

	l.Debug("Loading session...")
	err := s.ReloadSession(sess)
	if err != nil {
		return nil, err
	}

	l.Info("Session loaded successfully.")
	return sess, nil
}

//ReloadSession reloads the data in a session
func (s *RedisSessionManager) ReloadSession(session *Session) error {
	l := s.Logger.With(
		zap.String("operation", "ReloadSession"),
		zap.String("sessionID", session.ID),
	)

	lastUpdatedKey := GetLastUpdatedKey()

	l.Debug("Reloading session...")
	sessionKey := getSessionKey(session.ID)
	all, err := s.Client.HGetAll(sessionKey).Result()
	if err != nil {
		l.Error(
			"Reloading session failed.",
			zap.String("lastUpdatedKey", lastUpdatedKey),
			zap.String("sessionKey", sessionKey),
			zap.Error(err),
		)
		return err
	}
	if len(all) == 0 {
		err := &extensions.SessionNotFoundError{SessionID: session.ID}
		l.Error(
			"Session was not found!",
			zap.String("lastUpdatedKey", lastUpdatedKey),
			zap.String("sessionKey", sessionKey),
			zap.Error(err),
		)
		return err
	}

	session.data = map[string]interface{}{}

	for k, v := range all {
		if k == lastUpdatedKey {
			if lastUpdated, err := strconv.ParseInt(v, 10, 64); err == nil {
				session.LastUpdated = lastUpdated
			}
		}
		item, _ := extensions.Deserialize(v)
		session.data[k] = item
	}

	l.Info("Session reloaded successfully.")

	return nil
}

//ValidateSession indicates whether a session is valid or should be updated
func (s *RedisSessionManager) ValidateSession(session *Session) (bool, error) {
	l := s.Logger.With(
		zap.String("operation", "ValidateSession"),
		zap.String("sessionID", session.ID),
	)

	lastUpdatedKey := GetLastUpdatedKey()
	hashKey := getSessionKey(session.ID)
	l.Debug("Validating session...")
	ts, err := s.Client.HGet(hashKey, lastUpdatedKey).Result()
	if err != nil {
		l.Error("Could not validate session.", zap.Error(err))
		return false, err
	}

	timestamp, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		l.Error(
			"Could not validate session (invalid timestamp).",
			zap.String("timestamp", ts),
			zap.Error(err),
		)
		return false, err
	}

	isValid := timestamp == session.LastUpdated
	l.Info("Session validated successfully.", zap.Bool("isValid", isValid))

	return isValid, nil
}

//SetKey sets a key in the specified session
func (s *RedisSessionManager) SetKey(session *Session, key string, value interface{}) error {
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

	_, err = s.Client.HMSet(hashKey, map[string]string{
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
