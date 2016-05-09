package components

import (
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"net/http"
)

// TODO(spastorelli): Remove when multiple topics are supported.
const TestTopic = "Test"

func validateJWT(r *http.Request) bool {
	// TODO(spastorelli): Implement the origin check of the websocket connection using JWT.
	return true
}

var defaultUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     validateJWT,
}

// MsgHub handles the websocket connections to the DojoHub.
type MsgHub struct {
	running  bool
	channels map[string]*channel
}

// NewMsgHub creates a MsgHub.
func NewMsgHub() *MsgHub {
	return &MsgHub{
		running:  false,
		channels: make(map[string]*channel),
	}
}

// Run starts the channels for the registered DojoHub apps.
func (h *MsgHub) Run() {
	channel := newChannel(TestTopic)
	go channel.Serve()
	h.channels[channel.topic] = channel
	h.running = true
}

// ServeHTTP handles websocket connections to the Message Hub.
func (h *MsgHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.running {
		glog.Fatal("Dojo MessageHub is not running, can't serve requests.")
	}
	ws, err := defaultUpgrader.Upgrade(w, r, nil)
	if err != nil {
		glog.Errorf("Error while upgrading to websocket connection: %v\n", err)
		return
	}

	if err := r.ParseForm(); err != nil {
		glog.Errorf("Error parsing form: %s", err)
		return
	}
	clientId := r.Form.Get("cid")
	if clientId == "" {
		glog.Error("Could not get the clientId")
	}

	// TODO(spastorelli): Get the topic value from the JWT.
	hub, ok := h.channels[TestTopic]
	if !ok {
		glog.Error("Could not find the topic for the connection")
		return
	}

	conn := newClient(clientId, hub.publish, ws)
	hub.subscribe <- conn
	go conn.handleOutbound()
	go conn.handleInbound()
}

// channel handles messages that are published to its associated topic by
// dispatching them to its subscribers.
type channel struct {
	topic       string
	subscribers map[string]*client
	subscribe   chan *client
	unSubscribe chan *client
	publish     chan *clientMsg
}

// newChannel returns a new channel initialized with the provided topic.
func newChannel(t string) *channel {
	return &channel{
		topic:       t,
		subscribers: make(map[string]*client),
		subscribe:   make(chan *client),
		unSubscribe: make(chan *client),
		publish:     make(chan *clientMsg, 256),
	}
}

// addSubscriber adds the provided client to the map of subscribers.
func (ch *channel) addSubscriber(client *client) {
	ch.subscribers[client.id] = client
}

// removeSubscriber removes the provided client from the map of subscribers, closing
// the client SendToClient go channel.
func (ch *channel) removeSubscriber(client *client) {
	if _, ok := ch.subscribers[client.id]; ok {
		glog.Infof("Client %s is not subscribed to the Topic %s", client.id, ch.topic)
		return
	}
	delete(ch.subscribers, client.id)
	close(client.sendToClient)
}

// Serve starts the channel to handles its messages and subscribers.
func (ch *channel) Serve() {
	for {
		select {
		case client := <-ch.subscribe:
			ch.addSubscriber(client)
		case client := <-ch.unSubscribe:
			ch.removeSubscriber(client)
		case msg := <-ch.publish:
			for _, client := range ch.subscribers {
				if client.id != msg.clientId {
					client.sendToClient <- msg
				}
			}
		}
	}
}

// clientMsg wraps a websocket connection message for a given client.
type clientMsg struct {
	clientId string
	payload  []byte
}

// client handles inbound and outbound messages of an underlying websocket connection.
type client struct {
	id               string
	ws               *websocket.Conn
	publishToChannel chan *clientMsg
	sendToClient     chan *clientMsg
}

// NewClient creates a new client for a given id, channel and websocket connection.
func newClient(id string, pub chan *clientMsg, ws *websocket.Conn) *client {
	return &client{
		id:               id,
		ws:               ws,
		publishToChannel: pub,
		sendToClient:     make(chan *clientMsg, 256),
	}
}

// handleOutbound pulls messages from the channel and writes them back to the websocket connection.
func (c *client) handleOutbound() {
	for {
		chMsg, ok := <-c.sendToClient
		if !ok {
			if err := c.ws.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
				glog.Errorf("Error while sending close message to client %s: %v\n", c.id, err)
				return
			}
		}
		msg := chMsg.payload
		if err := c.ws.WriteMessage(websocket.TextMessage, msg); err != nil {
			// TODO(spastorelli): Unsubscribe the client if the error is websocket: close?
			glog.Errorf("Error while writing message '%v': %v\n", msg, err)
			return
		}
	}
}

// handleInbound reads messages from the websocket connection and publish them to the channel.
func (c *client) handleInbound() {
	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			glog.Errorf(
				"Error while reading message from the websocket: %v\n", err)
			break
		}
		clientMsg := &clientMsg{clientId: c.id, payload: msg}
		c.publishToChannel <- clientMsg
	}
}
