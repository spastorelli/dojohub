package components

import (
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"net/http"
)

type Status int8

const (
	Terminated Status = iota - 1
	Stopped
	Running
)

// MsgHub handles the websocket connections to the DojoHub.
type MsgHub struct {
	status     Status
	apps       map[string]*Application
	wsUpgrader websocket.Upgrader
}

// NewMsgHub creates a MsgHub.
func NewMsgHub() *MsgHub {
	hub := &MsgHub{
		status: Stopped,
		apps:   make(map[string]*Application),
	}
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     hub.validateClient,
	}
	hub.wsUpgrader = upgrader
	return hub
}

// validationKey returns the application secret key to validate the JWT token.
func (h *MsgHub) validationKey(token *jwt.Token) (key interface{}, err error) {
	// TODO(spastorelli): Define specific token validation errors.
	if jwt.SigningMethodHS256.Alg() != token.Header["alg"] {
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	}

	claimedAppId, ok := token.Claims["aud"].(string)
	if !ok {
		return nil, fmt.Errorf("No Application Id found in the token claims.")
	}

	app, ok := h.apps[claimedAppId]
	if !ok {
		return nil, fmt.Errorf("No corresponding Application found.")
	}

	return app.Secret()
}

// validateClient validates the client's JWT token,
func (h *MsgHub) validateClient(r *http.Request) bool {
	if err := r.ParseForm(); err != nil {
		glog.Errorf("Error while parsing the request parameters: %v", err)
		return false
	}

	token := r.Form.Get("t")
	if token == "" {
		glog.Error("Could not retrieve the token from request.")
		return false
	}

	parsedToken, err := jwt.Parse(token, h.validationKey)
	if err != nil {
		glog.Errorf("Error while validating the token: %v", err)
		return false
	}

	if parsedToken.Valid {
		cId := parsedToken.Claims["sub"].(string)
		appId := parsedToken.Claims["aud"].(string)
		// TODO(spastorelli): sub value has the format provider|user_id. Strip provider.
		r.Form.Set("cid", cId)
		r.Form.Set("aid", appId)
		return true
	}

	return false
}

// RegisterApplication registers an application with the MsgHub.
func (h *MsgHub) RegisterApplication(app *Application) {
	h.apps[app.Id] = app
}

// Run starts the channels for the registered applications.
func (h *MsgHub) Run() {
	for _, app := range h.apps {
		app.Serve()
	}
	h.status = Running
}

// ServeHTTP handles websocket client connections to the Message Hub.
func (h *MsgHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if Running != h.status {
		glog.Fatal("Dojo MessageHub is not running, can't serve requests.")
	}
	ws, err := h.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		glog.Errorf("Error while upgrading to websocket connection: %v\n", err)
		return
	}

	if err := r.ParseForm(); err != nil {
		glog.Errorf("Error parsing form: %s", err)
		return
	}
	clientId := r.Form.Get("cid")
	appId := r.Form.Get("aid")
	if clientId == "" || appId == "" {
		glog.Error("Could not get the client or application Id.")
	}

	app, ok := h.apps[appId]
	if !ok {
		glog.Error("Could not find an Application to connect the client to.")
		return
	}

	client := newClient(clientId, app.Channel.publish, ws)
	app.Channel.subscribe <- client
	if !(<-client.subscribed) {
		glog.Error("Could not subscribe client.")
	}
	go client.handleOutbound()
	go client.handleInbound()
}

// Application defines a DojoHub application.
type Application struct {
	Id           string
	Name         string
	b64EncSecret string
	Channel      *channel
	status       Status
}

// NewApplication creates a new Application instance.
func NewApplication(id string, name string, encSecret string) *Application {
	ch := newChannel(id)
	return &Application{id, name, encSecret, ch, Stopped}
}

// Serve start the application's channel to handle clients requests.
func (a *Application) Serve() {
	a.Channel.Serve()
	a.status = Running
}

// Secret returns the application decoded secret.
func (a *Application) Secret() (key []byte, err error) {
	return base64.URLEncoding.DecodeString(a.b64EncSecret)
}

// channel handles messages that are published to its associated topic by
// dispatching them to its subscribers.
type channel struct {
	topic       string
	subscribers map[string]*client
	subscribe   chan *client
	subAck      chan bool
	unSubscribe chan *client
	publish     chan *clientMsg
	terminate   chan bool
}

// newChannel returns a new channel initialized with the provided topic.
func newChannel(t string) *channel {
	return &channel{
		topic:       t,
		subscribers: make(map[string]*client),
		subscribe:   make(chan *client),
		subAck:      make(chan bool),
		unSubscribe: make(chan *client),
		publish:     make(chan *clientMsg, 256),
		terminate:   make(chan bool),
	}
}

// addSubscriber adds the provided client to the map of subscribers.
func (ch *channel) addSubscriber(client *client) {
	ch.subscribers[client.id] = client
}

// removeSubscriber removes the provided client from the map of subscribers, closing
// the client SendToClient go channel.
func (ch *channel) removeSubscriber(client *client) {
	if _, ok := ch.subscribers[client.id]; !ok {
		glog.Infof("Client %s is not subscribed to the Topic %s", client.id, ch.topic)
		return
	}
	delete(ch.subscribers, client.id)
	close(client.sendToClient)
}

// Serve starts the channel to handles its messages and subscribers.
func (ch *channel) Serve() {
	go func() {
		for {
			select {
			case client := <-ch.subscribe:
				ch.addSubscriber(client)
				client.subscribed <- true
			case client := <-ch.unSubscribe:
				ch.removeSubscriber(client)
				client.subscribed <- false
			case msg := <-ch.publish:
				for _, client := range ch.subscribers {
					if client.id != msg.clientId {
						client.sendToClient <- msg
					}
				}
			case <-ch.terminate:
				ch.Terminate()
				break
			}
		}
	}()
}

// Terminate terminates the channel.
func (ch *channel) Terminate() {
	for _, client := range ch.subscribers {
		ch.removeSubscriber(client)
	}
}

// wsConn is a websocket connection.
//
// Defining an interface that replicates the websocket.Conn type for testing purposes,
// since no Conn interface are defined in the gorilla/websocket package and there are
// no plans in defining one: https://github.com/gorilla/websocket/issues/74
type wsConn interface {
	WriteMessage(msgType int, data []byte) (err error)
	ReadMessage() (msgType int, payload []byte, err error)
}

// clientMsg wraps a websocket connection message for a given client.
type clientMsg struct {
	clientId string
	payload  []byte
}

// client handles inbound and outbound messages of an underlying websocket connection.
type client struct {
	id               string
	ws               wsConn
	publishToChannel chan *clientMsg
	sendToClient     chan *clientMsg
	subscribed       chan bool
}

// NewClient creates a new client for a given id, channel and websocket connection.
func newClient(id string, pub chan *clientMsg, ws wsConn) *client {
	return &client{
		id:               id,
		ws:               ws,
		publishToChannel: pub,
		sendToClient:     make(chan *clientMsg, 256),
		subscribed:       make(chan bool),
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
