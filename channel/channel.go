// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package channel

import (
	"os"
	"time"

	"github.com/heynemann/level/extensions/redis"
	"github.com/iris-contrib/middleware/logger"
	"github.com/iris-contrib/middleware/recovery"
	"github.com/kataras/iris"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
	redisCli "gopkg.in/redis.v4"
)

//Channel is responsible for communicating clients and backend servers
type Channel struct {
	Client        *redisCli.Client
	Config        *viper.Viper
	Logger        zap.Logger
	ServerOptions *Options
	WebApp        *iris.Framework
}

//New opens a new channel connection
func New(options *Options, logger zap.Logger) (*Channel, error) {
	if options == nil {
		options = DefaultOptions()
	}
	l := logger.With(
		zap.String("source", "channel"),
		zap.String("host", options.Host),
		zap.Int("port", options.Port),
		zap.Bool("debug", options.Debug),
	)
	c := Channel{
		Logger:        l,
		ServerOptions: options,
		Config:        viper.New(),
	}

	err := c.initializeChannel()
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Channel) initializeChannel() error {
	l := c.Logger.With(
		zap.String("operation", "initializeChannel"),
	)
	start := time.Now()
	l.Debug("Initializing channel...")

	c.setDefaultConfigurationOptions()
	err := c.initializeRedis()
	if err != nil {
		return err
	}
	//c.initializeNATS()
	c.initializeWebApp()

	l.Info(
		"Channel initialized successfully.",
		zap.Duration("channelInitialization", time.Now().Sub(start)),
	)

	return nil
}

func (c *Channel) setDefaultConfigurationOptions() {
	c.Config.SetDefault("channel.workingString", "WORKING")

	c.Config.SetDefault("services.redis.host", "localhost")
	c.Config.SetDefault("services.redis.port", 7777)
	c.Config.SetDefault("services.redis.password", "")
	c.Config.SetDefault("services.redis.db", 0)
}

func (c *Channel) initializeRedis() error {
	l := c.Logger.With(
		zap.String("operation", "initializeRedis"),
	)

	l.Debug("Initializing redis...")
	cli, err := redis.New(
		c.Config.GetString("services.redis.host"),
		c.Config.GetInt("services.redis.port"),
		c.Config.GetString("services.redis.password"),
		c.Config.GetInt("services.redis.db"),
		c.Logger,
	)
	if err != nil {
		l.Error("Initializing redis failed.", zap.Error(err))
		return err
	}
	l.Info("Redis initialized successfully.")
	c.Client = cli

	return nil
}

func (c *Channel) initializeWebApp() {
	debug := c.ServerOptions.Debug

	c.WebApp = iris.New()

	if debug {
		c.WebApp.Use(logger.New(iris.Logger))
	}
	c.WebApp.Use(recovery.New(os.Stderr))

	//a.Get("/healthcheck", HealthCheckHandler(app))

	//opt := cors.Options{AllowedOrigins: []string{"*"}}
	//a.Use(cors.New(opt)) // crs

	//SocketSupport(app)

	//app.connectNats()
	//app.connectRedis()
	//app.startMatchmaker()

}
