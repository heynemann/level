// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package channel

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/heynemann/level/messaging"
)

//NewWebSocketHandler creates a new websocket handler that sends and receives messages
func NewWebSocketHandler(channel *Channel) *WebsocketHandler {
	handler := &WebsocketHandler{
		Channel: channel,
		Mux:     map[string]func(http.ResponseWriter, *http.Request){},
	}

	handler.Mux["/"] = handler.ServeWS

	return handler
}

//WebsocketHandler struct
type WebsocketHandler struct {
	Channel *Channel
	Mux     map[string]func(http.ResponseWriter, *http.Request)
}

func (ws *WebsocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := ws.Mux[r.URL.String()]; ok {
		h(w, r)
		return
	}

	io.WriteString(w, "My server: "+r.URL.String())
}

//ServeWS is the / route that serves websocket connections
func (ws *WebsocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	c := ws.Channel

	var upgrader = websocket.Upgrader{} // use default options
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
		_, message, err := conn.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), "abnormal closure") {
				return
			}
			fmt.Println("read:", err)
			continue
		}
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
