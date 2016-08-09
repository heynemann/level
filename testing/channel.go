// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

// Source file: https://github.com/nats-io/nats/blob/master/test/test.go
// Copyright 2015 Apcera Inc. All rights reserved.

package testing

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/heynemann/level/channel"
	"github.com/heynemann/level/messaging"
	"github.com/onsi/gomega/types"
	"github.com/uber-go/zap"
)

// DefaultTestOptions returns default options for channel tests
func DefaultTestOptions() *channel.Options {
	return channel.NewOptions(
		"0.0.0.0",
		3000,
		true,
		"../config/test.yaml",
	)
}

//RunChannel will run a server on the default port.
func RunChannel(logger zap.Logger) (*channel.Channel, error) {
	return RunChannelOnPort(3000, logger)
}

//RunChannelOnPort will run a server on the given port.
func RunChannelOnPort(port int, logger zap.Logger) (*channel.Channel, error) {
	options := DefaultTestOptions()
	options.Port = port
	return RunChannelWithOptions(options, logger)
}

//RunChannelWithOptions allows running a very customized channel
func RunChannelWithOptions(options *channel.Options, logger zap.Logger) (*channel.Channel, error) {
	channel, err := channel.New(options, logger)
	if err != nil {
		return nil, err
	}

	go func() {
		channel.Start()
	}()
	return channel, nil
}

//TestConnection to a channel
type TestConnection struct {
	Channel    *channel.Channel
	ws         *websocket.Conn
	Received   []*messaging.Event
	Errors     []error
	Stop       chan bool
	PingTicker *time.Ticker
	Waited     int
	SessionID  string
}

//NewChannelTestConnection creates a new test connection to a channel
func NewChannelTestConnection(channel *channel.Channel) (*TestConnection, error) {
	timeout := time.Duration(100 * time.Millisecond)
	pongWait := 60 * time.Millisecond

	tc := &TestConnection{
		Channel:    channel,
		Received:   []*messaging.Event{},
		Errors:     []error{},
		Stop:       make(chan bool),
		PingTicker: time.NewTicker(50 * time.Millisecond),
	}
	wsURL := fmt.Sprintf("%s:%d", channel.ServerOptions.Host, channel.ServerOptions.Port)

	u := url.URL{Scheme: "ws", Host: wsURL, Path: "/"}

	start := time.Now()

	var ws *websocket.Conn
	var err error

	for {
		if time.Now().Sub(start) > timeout {
			return nil, fmt.Errorf("Timed out trying to establish websocket connection to %s.", wsURL)
		}

		ws, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			time.Sleep(5 * time.Millisecond)
			continue
		}
		break
	}
	tc.ws = ws

	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	go func(conn *TestConnection) {
		for {
			select {
			case <-conn.PingTicker.C:
				if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					return
				}
			case <-conn.Stop:
				return
			default:
				ev, err := conn.Receive()
				if err != nil {
					conn.Errors = append(conn.Errors, err)
					if strings.HasSuffix(err.Error(), "i/o timeout") {
						close(conn.Stop)
						return
					}
				}
				if ev == nil {
					continue
				}
				if ev.Key == "channel.session.started" {
					sid := ev.Payload.(map[string]interface{})["sessionID"]
					if sid != nil {
						conn.SessionID = sid.(string)
					}
				}
				conn.Received = append(conn.Received, ev)
			}
		}
	}(tc)

	return tc, nil
}

// Close websocket connection and stop listening for events.
func (tc *TestConnection) Close() {
	close(tc.Stop)
	tc.ws.Close()
}

//Send an action through the channel
func (tc *TestConnection) Send(action *messaging.Action) error {
	payload, _ := action.MarshalJSON()
	err := tc.ws.WriteMessage(websocket.TextMessage, payload)
	if err != nil {
		return err
	}

	return nil
}

//SendAction with the proper session ID
func (tc *TestConnection) SendAction(key string, payload interface{}) error {
	if tc.SessionID == "" {
		return fmt.Errorf("The sessionID for this connection is null. You should wait for the session started event before sending an action.")
	}
	action := messaging.NewAction(tc.SessionID, key, payload)
	payloadJSON, _ := action.MarshalJSON()
	err := tc.ws.WriteMessage(websocket.TextMessage, payloadJSON)
	if err != nil {
		return err
	}

	return nil
}

//WaitFor messages to come
func (tc *TestConnection) WaitFor(messages int, to ...time.Duration) error {
	timeout := 50 * time.Millisecond
	if to != nil && len(to) == 1 {
		timeout = to[0]
	}

	start := time.Now()

	for {
		if time.Now().Sub(start) > timeout {
			return fmt.Errorf("Timed out waiting for WebSocket messages to come in.")
		}

		if len(tc.Received) >= messages {
			tc.Waited += messages
			return nil
		}

		time.Sleep(5 * time.Millisecond)
	}
}

//WaitForEvent to come
func (tc *TestConnection) WaitForEvent(eventKey string, to ...time.Duration) (*messaging.Event, error) {
	timeout := 50 * time.Millisecond
	if to != nil && len(to) == 1 {
		timeout = to[0]
	}

	start := time.Now()

	for {
		if time.Now().Sub(start) > timeout {
			return nil, fmt.Errorf("Timed out waiting for Event %s to come in.", eventKey)
		}

		var foundEvent *messaging.Event
		for _, event := range tc.Received {
			if event.Key == eventKey {
				foundEvent = event
			}
		}
		if foundEvent != nil {
			return foundEvent, nil
		}

		time.Sleep(5 * time.Millisecond)
	}
}

//Receive the next event
func (tc *TestConnection) Receive(to ...time.Duration) (*messaging.Event, error) {
	c := tc.ws

	timeout := 10 * time.Millisecond
	if to != nil && len(to) == 1 {
		timeout = to[0]
	}
	c.SetReadDeadline(time.Now().Add(timeout))

	_, message, err := c.ReadMessage()
	if err != nil {
		return nil, err
	}
	ev := messaging.Event{}
	err = ev.UnmarshalJSON(message)
	if err != nil {
		return nil, err
	}

	return &ev, nil
}

//HaveEvent gomega matcher
func HaveEvent(eventKey string) types.GomegaMatcher {
	return &haveEventMatcher{
		eventKey: eventKey,
	}
}

type haveEventMatcher struct {
	eventKey string
}

func (matcher *haveEventMatcher) Match(actual interface{}) (success bool, err error) {
	client, ok := actual.(*TestConnection)
	if !ok {
		return false, fmt.Errorf("The HaveEvent matcher can only be used with *TestConnection instances.")
	}

	for _, event := range client.Received {
		if event.Key == matcher.eventKey {
			return true, nil
		}
	}

	return false, fmt.Errorf(
		"Event '%v' was not received yet.",
		matcher.eventKey,
	)
}

func (matcher *haveEventMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected event %s to have occurred.", matcher.eventKey)
}

func (matcher *haveEventMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected event %s not to have occurred.", matcher.eventKey)
}

//HavePayload gomega matcher
func HavePayload(payloadKey string) types.GomegaMatcher {
	return &havePayloadMatcher{
		payloadKey: payloadKey,
	}
}

type havePayloadMatcher struct {
	payloadKey string
}

func (matcher *havePayloadMatcher) Match(actual interface{}) (success bool, err error) {
	event, ok := actual.(*messaging.Event)
	if !ok {
		return false, fmt.Errorf("The HavePayload matcher can only be used with *messaging.Event instances.")
	}

	if _, ok := event.Payload.(map[string]interface{})[matcher.payloadKey]; ok {
		return true, nil
	}

	return false, fmt.Errorf(
		"Event '%s' does not have key '%s'.",
		event.Key,
		matcher.payloadKey,
	)
}

func (matcher *havePayloadMatcher) FailureMessage(actual interface{}) (message string) {
	event, _ := actual.(*messaging.Event)

	payloadJSON, _ := json.MarshalIndent(event.Payload, "", "  ")

	return fmt.Sprintf(
		"Expected event %s to have '%s' key in payload, but it had:\n %s",
		event.Key, matcher.payloadKey, payloadJSON,
	)
}

func (matcher *havePayloadMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	event, _ := actual.(*messaging.Event)
	return fmt.Sprintf("Expected event %s not to have '%s' key in payload, but it had.", event.Key, matcher.payloadKey)
}

//HavePayloadLike gomega matcher
func HavePayloadLike(payloadKey string, value interface{}) types.GomegaMatcher {
	return &havePayloadLikeMatcher{
		payloadKey: payloadKey,
		value:      value,
	}
}

type havePayloadLikeMatcher struct {
	payloadKey string
	value      interface{}
}

func (matcher *havePayloadLikeMatcher) Match(actual interface{}) (success bool, err error) {
	event, ok := actual.(*messaging.Event)
	if !ok {
		return false, fmt.Errorf("The HavePayloadLike matcher can only be used with *messaging.Event instances.")
	}

	if v, ok := event.Payload.(map[string]interface{})[matcher.payloadKey]; ok {
		if v == matcher.value {
			return true, nil
		}
	} else {
		return false, fmt.Errorf(
			"Event '%s' does not have key '%s'.",
			event.Key,
			matcher.payloadKey,
		)
	}

	return false, fmt.Errorf(
		"Event '%s' has key '%s', but the value '%v' does not match expected value of '%v'.",
		event.Key,
		matcher.payloadKey,
		event.Payload.(map[string]interface{})[matcher.payloadKey],
		matcher.value,
	)
}

func (matcher *havePayloadLikeMatcher) FailureMessage(actual interface{}) (message string) {
	event, _ := actual.(*messaging.Event)

	payloadJSON, _ := json.MarshalIndent(event.Payload, "", "  ")

	return fmt.Sprintf(
		"Expected event %s to have '%s' key in payload with value of '%v', but instead it has:\n %s",
		event.Key, matcher.payloadKey, matcher.value, payloadJSON,
	)
}

func (matcher *havePayloadLikeMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	event, _ := actual.(*messaging.Event)

	payloadJSON, _ := json.MarshalIndent(event.Payload, "", "  ")

	return fmt.Sprintf(
		"Expected event %s not to have '%s' key in payload with value of '%v', but it has:\n %s",
		event.Key, matcher.payloadKey, matcher.value, payloadJSON,
	)
}
