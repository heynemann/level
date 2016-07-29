// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package session

import (
	"fmt"

	"github.com/heynemann/level/extensions/serviceRegistry"
	"github.com/heynemann/level/extensions/sessionManager"
	"github.com/heynemann/level/messaging"
	"github.com/heynemann/level/metadata"
)

//Service represents the heartbeat service
type Service struct {
	ServiceID      string
	Registry       *registry.ServiceRegistry
	SessionManager sessionManager.SessionManager
}

//NewSessionService creates a new instance of a heartbeat service
func NewSessionService(serviceID string, serviceRegistry *registry.ServiceRegistry) (*Service, error) {
	s := &Service{
		ServiceID: serviceID,
		Registry:  serviceRegistry,
	}

	err := s.Initialize()
	if err != nil {
		return nil, err
	}

	return s, nil
}

//GetServiceDetails ditto
func (s *Service) GetServiceDetails() *registry.ServiceDetails {
	return registry.NewServiceDetails(
		s.ServiceID,
		"channel.session",
		"Session Manager",
		"Manages the session for all players in all games",
		metadata.VERSION,
		false,
	)
}

//GetServiceID returns the service ID for each instance of this service
func (s *Service) GetServiceID() string {
	return s.ServiceID
}

//HandleSessionRejoin happens when a player wants to rejoin the channel
func (s *Service) HandleSessionRejoin(sessionID string, action *messaging.Action) (*messaging.Event, error) {
	return nil, nil
}

//HandleSessionStart happens when a player first joins the channel
func (s *Service) HandleSessionStart(sessionID string, action *messaging.Action) (*messaging.Event, error) {
	event := messaging.NewEvent("channel.session.started", map[string]interface{}{"sessionID": sessionID})
	return event, nil
}

//HandleAction handles a given action for a player
func (s *Service) HandleAction(subject string, action *messaging.Action) (*messaging.Event, error) {
	switch action.Key {
	case "channel.session.start":
		return s.HandleSessionStart(action.SessionID, action)
	case "channel.session.rejoin":
		return s.HandleSessionRejoin(action.SessionID, action)
	default:
		return nil, fmt.Errorf("Cannot process action identified by: %s", action.Key)
	}
}

//Initialize the service - register it with the service registry
func (s *Service) Initialize() error {
	s.Registry.Register(s)
	return nil
}
