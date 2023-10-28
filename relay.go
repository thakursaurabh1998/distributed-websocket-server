package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Relay struct {
	Server        *Server
	Redis         *Redis
	ClientStore   *ClientStore
	LocalChannels *map[string]chan *Event
}

func (relay *Relay) Start() {
	log.Println("[*] Starting HTTP Server")
	go relay.Server.Listen()

	for client := range relay.Server.ClientConnected {
		relay.HandleConnection(client)
	}
}

func (relay *Relay) HandleConnection(client *Client) {
	relay.ClientStore.Lock()
	relay.ClientStore.Clients[client.ID] = client
	go client.Process(relay)

	log.Println("[*] Clients: ", len(relay.ClientStore.Clients))
	relay.ClientStore.Unlock()
}

func (relay *Relay) HandleIncoming(event *Event, client *Client) {
	switch event.Type {
	case "subscribe":
		if relay.Redis == nil {
			go client.Subscribe(relay, event.Channel)
		} else {
			go client.SubscribeRedis(relay, event.Channel)
		}
	case "message":
		if relay.Redis == nil {
			relay.Publish(event)
		} else {
			relay.PublishRedis(event)
		}
	default:
		log.Printf("%s", event)
	}
}

func (relay *Relay) Publish(event *Event) {
	for _, client := range relay.ClientStore.Clients {
		// uncomment when we have a client.Subscribed method
		if client.IsSubscribed(event.Channel) {
			client.outgoing <- event
		}
	}
}

func (relay *Relay) PublishRedis(event *Event) {
	if len(event.Channel) == 0 {
		return
	}

	payload, err := json.Marshal(event)

	if err != nil {
		log.Println("JSON encoding error", err)
		return
	}

	relay.Redis.Client.Publish(event.Channel, payload)
}

// Catch SIGHUP and SIGTERM for teardown processing
func (relay *Relay) Terminate() {
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGHUP, syscall.SIGTERM)

	go func() {
		<-signals
		os.Exit(0)
	}()
}
