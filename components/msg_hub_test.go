package components

import (
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
