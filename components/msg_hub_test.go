package components

import (
	"encoding/base64"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"net/url"
	"testing"
)

// mockWsConn mocks out a message that can be read from and written to
// a websocket connection.
type mockWsConnMsg struct {
	msgType int
	payload []byte
}

// mockWsConn mocks out a websocket connection.
type mockWsConn struct {
	messages chan *mockWsConnMsg
}

func newMockWsConn() *mockWsConn {
	return &mockWsConn{messages: make(chan *mockWsConnMsg)}
}

func (c *mockWsConn) WriteMessage(msgType int, data []byte) (err error) {
	m := &mockWsConnMsg{msgType, data}
	c.messages <- m
	return
}

func (c *mockWsConn) ReadMessage() (msgType int, data []byte, err error) {
	m := <-c.messages
	msgType = m.msgType
	data = m.payload
	return
}

func TestClientSubscribeToChannel(t *testing.T) {
	channel := newChannel("TestTopic")
	channel.Serve()

	expectedSubscriberCount := 0
	actualSubscriberCount := len(channel.subscribers)
	if len(channel.subscribers) != expectedSubscriberCount {
		t.Fatalf(
			"Channel subscribers count should be %d, got %d",
			expectedSubscriberCount, actualSubscriberCount)
	}

	clients := []string{"TestClient1", "TestClient2", "TestClient3", "TestClient4"}
	for _, client_id := range clients {
		ws := newMockWsConn()
		client := newClient(client_id, channel.publish, ws)
		channel.subscribe <- client
		<-client.subscribed
	}

	expectedSubscriberCount = 4
	actualSubscriberCount = len(channel.subscribers)
	if len(channel.subscribers) != expectedSubscriberCount {
		t.Fatalf(
			"Channel subscribers count should be %d, got %d",
			expectedSubscriberCount, actualSubscriberCount)
	}
	channel.terminate <- true
}

func TestClientSendMessages(t *testing.T) {
	channel := newChannel("TestTopic")
	channel.Serve()

	conns := map[string]*mockWsConn{
		"TestClient1": nil,
		"TestClient2": nil,
		"TestClient3": nil,
	}
	for id := range conns {
		ws := newMockWsConn()
		client := newClient(id, channel.publish, ws)
		channel.subscribe <- client
		<-client.subscribed
		go client.handleInbound()
		go client.handleOutbound()
		conns[id] = ws
	}

	expectedMsgSent := "hello world"

	conn1 := conns["TestClient1"]
	m1 := &mockWsConnMsg{1, []byte(expectedMsgSent)}
	conn1.messages <- m1

	receivers := []string{"TestClient2", "TestClient3"}

	for _, cId := range receivers {
		rConn := conns[cId]
		m := <-rConn.messages
		if m == nil {
			t.Fatal("Message not received by %s.", cId)
		}

		t.Logf("Message received by %s: %v", cId, m)

		actualMsgReceived := string(m.payload)
		if actualMsgReceived != expectedMsgSent {
			t.Fatalf(
				"Message received (%v) does not match message sent (%v)",
				actualMsgReceived, expectedMsgSent)
		}
	}
}

func createClientRequest(clientId, appId, secret string) (*http.Request, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims["aud"] = appId
	token.Claims["sub"] = clientId

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	f := url.Values{}
	f.Add("t", signedToken)
	return &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/ws"},
		Form:   f,
	}, nil
}

func TestClientValidation(t *testing.T) {
	h := NewMsgHub()

	// Register application
	appId := "testAppID"
	appName := "testAppName"
	appSecret := "testAppSecret"
	encSecret := base64.URLEncoding.EncodeToString([]byte(appSecret))
	h.RegisterApplication(NewApplication(appId, appName, encSecret))

	// Create a valid client
	clientId := "testClientID"
	validClientReq, err := createClientRequest(clientId, appId, appSecret)
	if err != nil {
		t.Fatalf("Error signing the Token: %v", err)
	}

	if !h.validateClient(validClientReq) {
		t.Fatalf("Expected client to be valid.")
	}

	claimedClientId := validClientReq.Form.Get("cid")
	if claimedClientId != clientId {
		t.Fatalf(
			"Client Id (%v) from claims does not match original Client Id (%v)",
			claimedClientId, clientId)
	}
	claimedAppId := validClientReq.Form.Get("aid")
	if claimedAppId != appId {
		t.Fatalf(
			"App Id (%v) from claims does not match original App Id (%v)",
			claimedAppId, appId)
	}

	// Create an invalid client using a bad secret
	invalidClientReq, err := createClientRequest("invalidClientID", appId, "badSecret")
	if err != nil {
		t.Fatalf("Error signing the Token: %v", err)
	}

	if h.validateClient(invalidClientReq) {
		t.Fatalf("Expected client to be invalid.")
	}
}
