// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package heartbeat

import (
	"fmt"
	"time"

	"github.com/uber-go/zap"

	"gopkg.in/redis.v4"
)

//Heartbeat extension responsible for service registry for all backend servers
type Heartbeat struct {
	ServerID           string
	RegistryExpiration time.Duration
	UpdateInterval     time.Duration
	Logger             zap.Logger
	client             *redis.Client
}

//NewDefault returns a new instance of the heartbeat extension with default options
func NewDefault(serverID string, redisHost string, redisPort int, redisPass string, redisDB int, logger zap.Logger) (*Heartbeat, error) {
	return New(serverID, redisHost, redisPort, redisPass, redisDB, logger, 3*time.Minute, 10*time.Second)
}

//New creates a new instance of the Heartbeat extension
func New(serverID, redisHost string, redisPort int, redisPass string, redisDB int, logger zap.Logger, registryExpiration, updateInterval time.Duration) (*Heartbeat, error) {
	l := logger.With(
		zap.String("source", "heartbeat"),
		zap.Duration("expiration", registryExpiration),
		zap.Duration("interval", updateInterval),
	)
	h := Heartbeat{
		ServerID:           serverID,
		RegistryExpiration: registryExpiration,
		UpdateInterval:     updateInterval,
		Logger:             l,
	}
	rl := l.With(
		zap.String("redisHost", redisHost),
		zap.Int("redisPort", redisPort),
		zap.Int("redisDB", redisDB),
	)

	rl.Debug("Connecting to Redis...")
	h.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
		Password: redisPass,
		DB:       redisDB,
	})

	_, err := h.client.Ping().Result()
	if err != nil {
		rl.Error("Could not connect to redis.", zap.Error(err))
		return nil, err
	}

	rl.Info("Connected to Redis successfully.")

	return &h, nil
}

//Register atomically registers a server with redis
func (h *Heartbeat) Register() error {
	registerServerScript := redis.NewScript(`
		local res
		res = redis.call("SET", KEYS[1], ARGV[1])
		res = redis.call("EXPIRE", KEYS[1], ARGV[2])
		res = redis.call("SADD", KEYS[2], ARGV[3])
		return res
	`)
	dt := fmt.Sprintf("%d", int32(time.Now().Unix()))
	serverStatusKey := fmt.Sprintf("server-status:%s", h.ServerID)
	_, err := registerServerScript.Run(
		h.client,
		[]string{serverStatusKey, "available-servers"},
		dt, int64(h.RegistryExpiration), h.ServerID,
	).Result()
	if err != nil {
		return err
	}

	return nil
}

//Start the server heartbeat
func (h *Heartbeat) Start() chan bool {
	done := make(chan bool)

	go func(self *Heartbeat) {
		for {
			select {
			case <-done:
				return
			default:
				err := self.Register()
				if err != nil {
					fmt.Println("Could not submit status to redis. Will retry in 10 seconds.", err)
				}
				fmt.Println("Status updated successfully in redis. Sleeping for 10 seconds...")
				time.Sleep(self.UpdateInterval)
			}
		}
	}(h)

	return done
}
