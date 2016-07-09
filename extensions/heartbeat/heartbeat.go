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

	"github.com/heynemann/level/extensions/redis"
	"github.com/uber-go/zap"

	redisCli "gopkg.in/redis.v4"
)

//Heartbeat extension responsible for service registry for all backend servers
type Heartbeat struct {
	ServerID           string
	RegistryExpiration time.Duration
	UpdateInterval     time.Duration
	Logger             zap.Logger
	Client             *redisCli.Client
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

	cli, err := redis.New(redisHost, redisPort, redisPass, redisDB, l)
	if err != nil {
		return nil, err
	}
	h.Client = cli

	return &h, nil
}

//Register atomically registers a server with redis
func (h *Heartbeat) Register() error {
	l := h.Logger.With(zap.String("operation", "Register"))
	registerServerScript := redisCli.NewScript(`
		local res
		res = redis.call("SET", KEYS[1], ARGV[1])
		res = redis.call("EXPIRE", KEYS[1], ARGV[2])
		res = redis.call("SADD", KEYS[2], ARGV[3])
		return res
	`)
	dt := fmt.Sprintf("%d", int32(time.Now().Unix()))
	serverStatusKey := fmt.Sprintf("server-status:%s", h.ServerID)

	start := time.Now()
	l.Debug("Registering server with service registry...")
	_, err := registerServerScript.Run(
		h.Client,
		[]string{serverStatusKey, "available-servers"},
		dt, int64(h.RegistryExpiration), h.ServerID,
	).Result()
	if err != nil {
		l.Error("Could not register with service registry.", zap.Error(err))
		return err
	}
	l.Info("Registered with service registry successfully.", zap.Duration("registerDuration", time.Now().Sub(start)))

	return nil
}

//Start the server heartbeat
func (h *Heartbeat) Start() chan bool {
	l := h.Logger.With(zap.String("operation", "Start"))
	done := make(chan bool)

	l.Debug("Starting heartbeat...")
	go func(self *Heartbeat) {
		for {
			select {
			case <-done:
				l.Debug("Stopping heartbeat...")
				return
			default:
				err := self.Register()
				if err == nil {
					l.Info("Status updated successfully in redis.")
				}
				time.Sleep(self.UpdateInterval)
			}
		}
	}(h)

	return done
}
