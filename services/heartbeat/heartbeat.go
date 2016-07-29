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
	"github.com/heynemann/level/metadata"
)

//Service represents the heartbeat service
type Service struct {
	ServiceID string
	Registry  *registry.ServiceRegistry
	PubSub    *pubsub.PubSub
}

//NewHeartbeatService creates a new instance of a heartbeat service
func NewHeartbeatService(serviceID string, serviceRegistry *registry.ServiceRegistry) (*Service, error) {
	s := &Service{
		ServiceID: serviceID,
		Registry:  serviceRegistry,
	}

	return s, nil
}

//GetServiceID returns the service ID for each instance of this service
func (s *Service) GetServiceID() string {
	return s.ServiceID
}

//GetServiceDetails ditto
func (s *Service) GetServiceDetails() *registry.ServiceDetails {
	return registry.NewServiceDetails(
		s.ServiceID,
		"channel.heartbeat",
		"Heartbeat",
		"Manages the session for all players in all games",
		metadata.VERSION,
		false,
	)
}

//HandleAction handles a given action for an user
func (s *Service) HandleAction(subject string, action *messaging.Action) (*messaging.Event, error) {
	switch action.Payload.(type) {
	case map[string]interface{}:
		event := messaging.NewEvent("channel.heartbeat.received", map[string]interface{}{
			"clientSent": action.Payload.(map[string]interface{})["clientSent"],
			"serverSent": time.Now().UnixNano() / 1000000,
		})

		return event, nil
	default:
		return nil, fmt.Errorf("Could not understand heartbeat payload: %v", action.Payload)
	}
}
