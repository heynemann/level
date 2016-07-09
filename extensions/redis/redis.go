// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package redis

import (
	"fmt"
	"time"

	"github.com/uber-go/zap"

	redisCli "gopkg.in/redis.v4"
)

//New creates a new connection to Redis
func New(redisHost string, redisPort int, redisPass string, redisDB int, logger zap.Logger) (*redisCli.Client, error) {
	l := logger.With(
		zap.String("operation", "New"),
		zap.String("redisHost", redisHost),
		zap.Int("redisPort", redisPort),
		zap.Int("redisDB", redisDB),
	)

	l.Debug("Connecting to Redis...")
	client := redisCli.NewClient(&redisCli.Options{
		Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
		Password: redisPass,
		DB:       redisDB,
	})

	start := time.Now()
	_, err := client.Ping().Result()
	if err != nil {
		l.Error("Could not connect to redis.", zap.Error(err))
		return nil, err
	}
	l.Info("Connected to Redis successfully.", zap.Duration("connection", time.Now().Sub(start)))

	return client, nil
}
