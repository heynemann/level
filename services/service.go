// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package service

import (
	"fmt"
	"os"

	"github.com/heynemann/level/extensions/serviceRegistry"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

var logLevels = map[string]int{
	"debug": int(zap.DebugLevel),
	"info":  int(zap.InfoLevel),
	"warn":  int(zap.WarnLevel),
	"error": int(zap.ErrorLevel),
	"panic": int(zap.PanicLevel),
	"fatal": int(zap.FatalLevel),
}

//Service describes a server interface
type Service interface {
	SetServerFlags(cmd *cobra.Command)
	SetDefaultConfigurations(*viper.Viper)
}

//Server identifies a service Server
type Server struct {
	Logger          zap.Logger
	ConfigPath      string
	Config          *viper.Viper
	ServiceRegistry *registry.ServiceRegistry
	Service         registry.Service
	ServerService   Service
	LogLevel        string
	Quit            chan bool
}

//NewServer creates an configures a new server instance
func NewServer(serv registry.Service, logger zap.Logger, configPath string) (*Server, error) {
	var service Service
	var ok bool
	if service, ok = serv.(Service); !ok {
		return nil, fmt.Errorf(
			"Service %s does not implement interface Service. Please refer to the docs.",
			serv.GetServiceDetails().Name,
		)
	}

	s := &Server{Service: serv, ServerService: service, Logger: logger, ConfigPath: configPath}
	err := s.Configure()
	if err != nil {
		return nil, err
	}

	return s, nil
}

//Configure the server
func (s *Server) Configure() error {
	s.Config = viper.New()
	s.SetDefaultConfiguration()
	s.ServerService.SetDefaultConfigurations(s.Config)

	s.LoadConfiguration(s.ConfigPath)
	err := s.initializeServiceRegistry()
	return err
}

//SetDefaultConfiguration options
func (s *Server) SetDefaultConfiguration() {
	s.Logger.Debug("Setting default configuration")
	s.Config.SetDefault("services.nats.url", "nats://localhost:4222")

	s.Config.SetDefault("services.redis.host", "localhost")
	s.Config.SetDefault("services.redis.port", 4444)
}

//LoadConfiguration from filesystem
func (s *Server) LoadConfiguration(configPath string) {
	if configPath == "" {
		s.Logger.Panic("Could not load configuration due to empty config path.")
		os.Exit(-1)
	}

	s.Config.SetConfigFile(configPath)
	s.Config.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := s.Config.ReadInConfig(); err == nil {
		s.Logger.Info("Loaded configuration file.", zap.String("configPath", s.Config.ConfigFileUsed()))
	}
}

func (s *Server) initializeServiceRegistry() error {
	natsURL := s.Config.GetString("services.nats.URL")
	l := s.Logger.With(
		zap.String("operation", "initializeServiceRegistry"),
		zap.String("natsURL", natsURL),
	)

	l.Debug("Initializing registry...")
	sr, err := registry.NewServiceRegistry(
		natsURL,
		s.Logger,
	)
	if err != nil {
		l.Error("Error initializing service registry.", zap.Error(err))
		return err
	}

	l.Info("Service registry initialized successfully.")
	s.ServiceRegistry = sr

	l.Debug("Registering service...")
	s.ServiceRegistry.Register(s.Service)
	l.Info("Service registered successfully.")

	return nil
}

//Listen to incoming messages
func (s *Server) Listen() {
	s.Logger.Info("Service listening for messages...")
	s.Quit = make(chan bool)

	for {
		select {
		case <-s.Quit:
			return
		}
	}
}

//Close the server running
func (s *Server) Close() {
	s.Quit <- true
}

func (s *Server) setServerFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&s.ConfigPath, "config", "c", "./config/local.yaml", "configuration file to initialize this server with")
	cmd.PersistentFlags().StringVarP(&s.LogLevel, "loglevel", "l", "warn", "default log level for this backend server")
}

func getCommandFor(s *Server) *cobra.Command {
	details := s.Service.GetServiceDetails()

	return &cobra.Command{
		Use:   details.Name,
		Short: details.Description,
		Long:  details.Description,

		Run: func(cmd *cobra.Command, args []string) {
			s.Logger = s.Logger.With(
				zap.String("serverName", details.Name),
				zap.String("serverDescription", details.Description),
				zap.String("serverVersion", details.Version),
				zap.String("serverID", details.ServiceID),
			)

			s.setServerFlags(cmd)
			s.ServerService.SetServerFlags(cmd)

			s.Logger.Debug("Running backend server")
			s.Listen()
		},
	}
}

//RunMultipleServices in a single server
func RunMultipleServices(logger zap.Logger, configPath string, services ...registry.Service) error {
	if len(services) == 0 {
		return fmt.Errorf("Can't configure server with no services.")
	}
	serv := services[0]

	s, err := NewServer(serv, logger, configPath)
	if err != nil {
		logger.Error("Backend server finalized with error!", zap.Error(err))
		os.Exit(-1)
	}

	for i := 1; i < len(services); i++ {
		s.ServiceRegistry.Register(services[i])
	}

	cmd := getCommandFor(s)
	if err = cmd.Execute(); err != nil {
		s.Logger.Error("Backend server finalized with error!", zap.Error(err))
		os.Exit(-1)
	}

	return nil
}

//Run a new Service
func Run(serv registry.Service, logger zap.Logger, configPath string) error {
	return RunMultipleServices(logger, configPath, serv)
}
