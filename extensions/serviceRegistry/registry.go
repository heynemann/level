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

//Service describes a service
type Service interface {
	GetServiceDetails() *ServiceDetails
	HandleAction(string, *messaging.Action) (*messaging.Event, error)
}

//ServiceDetails identify a service
type ServiceDetails struct {
	ServiceID   string
	Namespace   string
	Name        string
	Description string
	Version     string
	Sticky      bool
}

//NewServiceDetails returns a new service details instance
func NewServiceDetails(serviceID, namespace, name, description, version string, sticky bool) *ServiceDetails {
	return &ServiceDetails{
		ServiceID:   serviceID,
		Namespace:   namespace,
		Name:        name,
		Description: description,
		Version:     version,
		Sticky:      sticky,
	}
}

//ServiceRegistry is the registry where all services specify their properties
type ServiceRegistry struct {
	Logger zap.Logger
	Client *nats.Conn
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

	rsr.Client = nc

	return rsr, nil
}

//Register a given service with the registry
func (s *ServiceRegistry) Register(service Service) error {
	s.listenForMessages(service)

	return nil
}

func (s *ServiceRegistry) listenForMessages(service Service) {
	l := s.Logger.With(
		zap.String("operation", "listenForMessages"),
	)

	l.Debug("Getting service details...")
	details := service.GetServiceDetails()

	var queue string
	if details.Sticky {
		queue = fmt.Sprintf("%s.%s", details.Namespace, details.ServiceID)
	} else {
		queue = fmt.Sprintf("%s.>", details.Namespace)
	}

	l.Debug(
		"Service details retrieved.",
		zap.Bool("sticky", details.Sticky),
		zap.String("queue", queue),
	)

	s.Client.QueueSubscribe(queue, "default", func(msg *nats.Msg) {
		action := messaging.Action{}
		action.UnmarshalJSON(msg.Data)

		la := l.With(
			zap.String("actionKey", action.Key),
			zap.String("action", string(msg.Data)),
		)
		la.Debug("Received action.")

		la.Debug("Handling action...")
		event, err := service.HandleAction(msg.Subject, &action)
		if err != nil {
			la.Error("Error handling action.", zap.Error(err))
			return
		}
		if event == nil {
			la.Warn("Message handling did not return an event.")
			return
		}
		l.Debug("Action handled.")
		eventJSON, err := event.MarshalJSON()
		if err != nil {
			la.Error("Error marshaling event.", zap.Error(err))
			return
		}
		s.Client.Publish(msg.Reply, eventJSON)
		la.Debug("Published event successfully.", zap.String("replyQueue", msg.Reply))
	})
}

//Terminate the service registry
func (s *ServiceRegistry) Terminate() {
	s.Client.Close()
}
