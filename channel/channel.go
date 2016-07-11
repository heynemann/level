// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package channel

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/heynemann/level/extensions/pubsub"
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
	Redis         *redisCli.Client
	Config        *viper.Viper
	Logger        zap.Logger
	PubSub        *pubsub.PubSub
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

	err := c.loadConfiguration()
	if err != nil {
		return err
	}

	err = c.initializeRedis()
	if err != nil {
		return err
	}

	err = c.initializePubSub()
	if err != nil {
		return err
	}

	c.initializeWebApp()

	l.Info(
		"Channel initialized successfully.",
		zap.Duration("channelInitialization", time.Now().Sub(start)),
	)

	return nil
}

func (c *Channel) setDefaultConfigurationOptions() {
	c.Config.SetDefault("channel.workingText", "WORKING")

	c.Config.SetDefault("channel.services.redis.host", "localhost")
	c.Config.SetDefault("channel.services.redis.port", 7777)
	c.Config.SetDefault("channel.services.redis.password", "")
	c.Config.SetDefault("channel.services.redis.db", 0)

	c.Config.SetDefault("channel.services.nats.URL", "nats://localhost:7778")
}

func (c *Channel) loadConfiguration() error {
	l := c.Logger.With(
		zap.String("operation", "loadConfiguration"),
		zap.String("configFile", c.ServerOptions.ConfigFile),
	)

	absConfigFile, err := filepath.Abs(c.ServerOptions.ConfigFile)
	if err != nil {
		l.Error("Configuration file not found.", zap.Error(err))
		return err
	}

	l = l.With(
		zap.String("absConfigFile", absConfigFile),
	)

	l.Info("Loading configuration.")

	if _, err := os.Stat(absConfigFile); os.IsNotExist(err) {
		l.Error("Configuration file not found.", zap.Error(err))
		return err
	}

	c.Config.SetConfigFile(c.ServerOptions.ConfigFile)
	c.Config.SetEnvPrefix("level") // read in environment variables that match
	c.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.Config.AutomaticEnv()

	// If a config file is found, read it in.
	if err := c.Config.ReadInConfig(); err != nil {
		l.Error("Configuration could not be loaded.", zap.Error(err))
		return err
	}

	l.Info(
		"Configuration loaded successfully.",
		zap.String("configPath", c.Config.ConfigFileUsed()),
	)
	return nil
}

func (c *Channel) initializeRedis() error {
	cli, err := redis.New(
		c.Config.GetString("channel.services.redis.host"),
		c.Config.GetInt("channel.services.redis.port"),
		c.Config.GetString("channel.services.redis.password"),
		c.Config.GetInt("channel.services.redis.db"),
		c.Logger,
	)
	if err != nil {
		return err
	}
	c.Redis = cli

	return nil
}

func (c *Channel) initializePubSub() error {
	pubsub, err := pubsub.New(c.Config.GetString("channel.services.nats.URL"), c.Logger)
	if err != nil {
		return err
	}
	c.PubSub = pubsub
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
