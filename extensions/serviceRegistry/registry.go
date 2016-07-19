// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package registry

import (
	"fmt"

	"github.com/heynemann/level/messaging"
	"github.com/nats-io/nats"
	"github.com/uber-go/zap"
)

//Action represents an action that can be handled by the Service
type Action struct {
	Key    string
	Sticky bool
}

//Service describes a service
type Service interface {
	//Initialize(*PubSub)
	GetServiceID() string
	GetServiceActions() []*Action
	HandleAction(string, *messaging.Action) (*messaging.Event, error)
}

//ServiceRegistry is the registry where all services specify their properties
type ServiceRegistry struct {
	Logger zap.Logger
	Client *nats.EncodedConn
}

//NewServiceRegistry returns a connected redis service registry
func NewServiceRegistry(natsURL string, logger zap.Logger) (*ServiceRegistry, error) {
	l := logger.With(
		zap.String("source", "serviceRegistry"),
		zap.String("natsURL", natsURL),
	)

	rsr := &ServiceRegistry{
		Logger: l,
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}

	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		return nil, err
	}

	rsr.Client = c

	return rsr, nil
}

//Register a given service with the registry
func (s *ServiceRegistry) Register(service Service) error {
	actions := service.GetServiceActions()

	for _, action := range actions {
		s.listenForMessages(service, action)
	}

	return nil
}

func (s *ServiceRegistry) listenForMessages(service Service, action *Action) {
	serviceID := service.GetServiceID()
	queue := action.Key
	if action.Sticky {
		queue = fmt.Sprintf("%s.%s", queue, serviceID)
	} else {
		queue = fmt.Sprintf("%s.>", queue)
	}

	s.Client.QueueSubscribe(queue, "default", func(subj, reply string, msg *nats.Msg) {
		action := messaging.Action{}
		action.UnmarshalJSON(msg.Data)
		event, err := service.HandleAction(subj, &action)
		if err != nil {
			//TODO: LOG ERROR
		}
		s.Client.Publish(reply, event)
	})
}

//Terminate the service registry
func (s *ServiceRegistry) Terminate() {
	s.Client.Close()
}
