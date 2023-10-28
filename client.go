package main

import (
	"bytes"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeTimeout   = 10 * time.Second
	maxMessageSize = 512
	pongPeriod     = 5 * time.Second
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	ID                 string
	Token              string
	SubscribedChannels map[string]bool
	conn               *websocket.Conn
	incoming           chan *Event
	outgoing           chan *Event
}

type ClientStore struct {
	sync.Mutex
	Clients map[string]*Client
}

func NewClientStore() *ClientStore {
	return &ClientStore{
		Clients: make(map[string]*Client),
	}
}

func NewClient(conn *websocket.Conn, token string) *Client {
	return &Client{
		ID:                 uuid.NewString(),
		Token:              token,
		SubscribedChannels: make(map[string]bool),
		conn:               conn,
		incoming:           make(chan *Event),
		outgoing:           make(chan *Event),
	}
}

func (client *Client) Process(relay *Relay) {
	go client.Heartbeat(relay)
	go client.Reader(relay)
	go client.Writer(relay)
	go client.Router(relay)
}

// Heartbeat tests to see if client remains conencted through passive polling
func (client *Client) Heartbeat(relay *Relay) {
	defer client.Close(relay)

	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(pongPeriod))
		return nil
	})

	for {
		err := client.conn.WriteMessage(websocket.PingMessage, []byte("keep-alive"))
		if websocket.IsCloseError(err) {
			return
		}

		client.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		time.Sleep(pongPeriod / 2)
	}
}

func (client *Client) Close(relay *Relay) {
	relay.ClientStore.Lock()
	client.conn.Close()
	delete(relay.ClientStore.Clients, client.ID)
	log.Println("Clients: ", len(relay.ClientStore.Clients))
	relay.ClientStore.Unlock()
}

func (client *Client) Reader(relay *Relay) {
	defer client.Close(relay)

	client.conn.SetReadDeadline(time.Now().Add(pongPeriod))
	client.conn.SetReadLimit(maxMessageSize)

	for {
		_, message, err := client.conn.ReadMessage()

		if err != nil {
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		event := &Event{}
		json.Unmarshal(message, event)
		client.incoming <- event
	}
}

func (client *Client) Writer(relay *Relay) {
	defer client.Close(relay)

	for event := range client.outgoing {
		client.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		writer, err := client.conn.NextWriter(websocket.TextMessage)

		if err != nil {
			return
		}

		payload, err := json.Marshal(event)

		if err != nil {
			log.Println("JSON encoding error", err)
			return
		}

		writer.Write(payload)

		if err := writer.Close(); err != nil {
			return
		}
	}
}

func (client *Client) Router(relay *Relay) {
	for event := range client.incoming {
		relay.HandleIncoming(event, client)
	}
}

func (client *Client) SubscribeRedis(relay *Relay, channel string) {
	client.addChannelToSubscribed(channel)

	psc := relay.Redis.Client.Subscribe(channel)
	defer psc.Close()

	for {
		message, err := psc.ReceiveMessage()

		if err != nil {
			panic(err)
		}

		event := &Event{}

		json.Unmarshal([]byte(message.Payload), event)
		client.outgoing <- event
	}
}

func (client *Client) Subscribe(relay *Relay, channel string) {
	client.addChannelToSubscribed(channel)
	for event := range client.incoming {
		client.outgoing <- event
	}
}

func (client *Client) IsSubscribed(channel string) bool {
	return client.SubscribedChannels[channel]
}

func (client *Client) addChannelToSubscribed(channel string) {
	client.SubscribedChannels[channel] = true
}
