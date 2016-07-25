// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package channel

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/heynemann/level/extensions/pubsub"
	"github.com/heynemann/level/extensions/redis"
	"github.com/heynemann/level/extensions/serviceRegistry"
	"github.com/heynemann/level/extensions/sessionManager"
	"github.com/heynemann/level/messaging"
	"github.com/heynemann/level/services/heartbeat"
	"github.com/kataras/iris"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
	redisCli "gopkg.in/redis.v4"
)

//Channel is responsible for communicating clients and backend servers
type Channel struct {
	Redis           *redisCli.Client
	SessionManager  sessionManager.SessionManager
	Config          *viper.Viper
	Logger          zap.Logger
	ServiceRegistry *registry.ServiceRegistry
	PubSub          *pubsub.PubSub
	ServerOptions   *Options
	WebApp          *iris.Framework
	SocketApp       string
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

	err = c.initializeSessionManager()
	if err != nil {
		return err
	}

	err = c.initializePubSub()
	if err != nil {
		return err
	}

	err = c.initializeServiceRegistry()
	if err != nil {
		return err
	}

	err = c.initializeDefaultServices()
	if err != nil {
		return err
	}

	c.initializeWebApp()
	c.initializeWebSocket()

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

	c.Config.SetDefault("channel.actionTimeout", 5)
	c.Config.SetDefault("channel.services.sessionManager.expiration", 180)
}

func (c *Channel) loadConfiguration() error {
	l := c.Logger.With(
		zap.String("operation", "loadConfiguration"),
		zap.String("configFile", c.ServerOptions.ConfigFile),
	)

	absConfigFile, _ := filepath.Abs(c.ServerOptions.ConfigFile)
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

func (c *Channel) initializeSessionManager() error {
	sm, err := sessionManager.GetRedisSessionManager(
		c.Config.GetString("channel.services.redis.host"),
		c.Config.GetInt("channel.services.redis.port"),
		c.Config.GetString("channel.services.redis.password"),
		c.Config.GetInt("channel.services.redis.db"),
		c.Config.GetInt("channel.services.sessionManager.expiration"),
		c.Logger,
	)
	if err != nil {
		return err
	}

	c.SessionManager = sm
	return nil
}

func (c *Channel) initializePubSub() error {
	pubsub, err := pubsub.New(
		c.Config.GetString("channel.services.nats.URL"),
		c.Logger,
		c.SessionManager,
		time.Duration(c.Config.GetInt("channel.actionTimeout"))*time.Second,
	)
	if err != nil {
		return err
	}
	c.PubSub = pubsub
	return nil
}

func (c *Channel) initializeServiceRegistry() error {
	sr, err := registry.NewServiceRegistry(
		c.Config.GetString("channel.services.nats.URL"),
		c.Logger,
	)
	if err != nil {
		return err
	}
	c.ServiceRegistry = sr
	return nil
}

func (c *Channel) initializeDefaultServices() error {
	hbID := fmt.Sprintf("services:heartbeat:%s", uuid.NewV4().String())
	_, err := heartbeat.NewHeartbeatService(hbID, c.ServiceRegistry)
	if err != nil {
		return err
	}
	//sessions, err := session.NewSessionService()
	//if err != nil {
	//return err
	//}

	return nil
}

//func (c *Channel) initializeWebApp() {
////debug := c.ServerOptions.Debug

//conf := config.Iris{
//DisableBanner: true,
//}

//c.WebApp = iris.New(conf)

////if debug {
////c.WebApp.Use(logger.New(iris.Logger))
////}
//c.WebApp.Use(recovery.New(os.Stderr))

//opt := cors.Options{AllowedOrigins: []string{"*"}}
//c.WebApp.Use(cors.New(opt)) // crs

//c.WebApp.Get("/healthcheck", HealthCheckHandler(c))
//}

//func (c *Channel) initializeWebSocket() {
//c.Logger.Debug("Initializing websocket...")
//c.WebApp.Config.Websocket.Endpoint = "/"
//ws := c.WebApp.Websocket // get the websocket server
//ws.OnConnection(func(socket iris.WebsocketConnection) {
//fmt.Println("got player connection")
//c.PubSub.RegisterPlayer(socket)
//c.PubSub.BindEvents(socket)
////socket.OnMessage(func(message []byte) {
////fmt.Println(string(message))
////})
//})
//c.Logger.Info("Websocket initialized.")
//}

func (c *Channel) initializeWebApp() {

}

func (c *Channel) initializeWebSocket() {
	var upgrader = websocket.Upgrader{} // use default options

	var handler = func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}

		player, err := c.PubSub.RegisterPlayer(conn)
		if err != nil {
			fmt.Println("Error registering player: ", err)
			return
		}

		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				if strings.Contains(err.Error(), "abnormal closure") {
					return
				}
				fmt.Println("read:", err)
				continue
			}
			fmt.Printf("recv: [%d] %s\n", mt, message)
			action := &messaging.Action{}
			err = action.UnmarshalJSON(message)
			if err != nil {
				fmt.Println("Error with action: ", err)
				continue
			}
			c.PubSub.RequestAction(player, action, func(event *messaging.Event) error {
				json, err := event.MarshalJSON()
				if err != nil {
					fmt.Println("Error sending action reply")
					return err
				}
				conn.WriteMessage(websocket.TextMessage, json)
				return nil
			})
		}
	}

	http.HandleFunc("/", handler)
}

//Start the channel
func (c *Channel) Start() {
	server := fmt.Sprintf("%s:%d", c.ServerOptions.Host, c.ServerOptions.Port)
	http.ListenAndServe(server, nil)
	//c.WebApp.Listen(server)
}
