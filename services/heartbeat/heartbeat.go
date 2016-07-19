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

	"github.com/heynemann/level/extensions/pubsub"
	"github.com/heynemann/level/extensions/serviceRegistry"
	"github.com/heynemann/level/messaging"
)

//Service represents the heartbeat service
type Service struct {
	ID       string
	Registry *registry.ServiceRegistry
	PubSub   *pubsub.PubSub
}

//NewHeartbeatService creates a new instance of a heartbeat service
func NewHeartbeatService(id string, serviceRegistry *registry.ServiceRegistry) (*Service, error) {
	s := &Service{
		ID:       id,
		Registry: serviceRegistry,
	}

	err := s.Initialize()
	if err != nil {
		return nil, err
	}

	return s, nil
}

//GetServiceID returns the service ID for each instance of this service
func (s *Service) GetServiceID() string {
	return s.ID
}

//Initialize the service - register it with the service registry
func (s *Service) Initialize() error {
	s.Registry.Register(s)
	return nil
}

//GetServiceActions returns all the actions the service can handle
func (s *Service) GetServiceActions() []*registry.Action {
	return []*registry.Action{
		&registry.Action{
			Key:    "channel.heartbeat",
			Sticky: false,
		},
	}
}

//HandleAction handles a given action for an user
func (s *Service) HandleAction(subject string, action *messaging.Action) (*messaging.Event, error) {
	switch action.Payload.(type) {
	case map[string]interface{}:
		event := messaging.NewEvent("channel.heartbeat", map[string]interface{}{
			"clientSent": action.Payload.(map[string]interface{})["clientSent"],
			"serverSent": time.Now().UnixNano() / 1000000,
		})

		return event, nil
	default:
		return nil, fmt.Errorf("Could not understand heartbeat payload: %v", action.Payload)
	}
}
