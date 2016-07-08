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

	"gopkg.in/redis.v4"
)

//Heartbeat extension responsible for service registry for all backend servers
type Heartbeat struct {
	ServerID           string
	RegistryExpiration time.Duration
	client             *redis.Client
}

//New creates a new instance of the Heartbeat extension
func New(serverID, redisHost string, redisPort int, redisPass string, redisDB int, registryExpiration time.Duration) (*Heartbeat, error) {
	h := Heartbeat{
		ServerID:           serverID,
		RegistryExpiration: registryExpiration,
	}
	h.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
		Password: redisPass,
		DB:       redisDB,
	})

	_, err := h.client.Ping().Result()
	if err != nil {
		return nil, err
	}

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
func (h *Heartbeat) Start() *chan bool {
	done := make(chan bool)

	go func(self *Heartbeat) {
		for {
			switch {
			case <-done:
				return
			default:
				err := self.Register()
				if err != nil {
					fmt.Println("Could not submit status to redis. Will retry in 10 seconds.", err)
				}
				fmt.Println("Status updated successfully in redis. Sleeping for 10 seconds...")
				time.Sleep(10 * time.Second)
			}
		}
	}(h)

	return &done
}
