package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Relay struct {
	Server      *Server
	Redis       *Redis
	ClientStore *ClientStore
}

func (relay *Relay) Start() {
	log.Println("Starting HTTP Server")
	go relay.Server.Listen()

	for {
		select {
		case client := <-relay.Server.ClientConnected:
			relay.HandleConnection(client)
			break
		}
	}
}

func (relay *Relay) HandleConnection(client *Client) {
	relay.ClientStore.Lock()
	relay.ClientStore.Clients[client.ID] = client
	go client.Process(relay)

	log.Println("Clients: ", len(relay.ClientStore.Clients))
	relay.ClientStore.Unlock()
}

func (relay *Relay) HandleIncoming(event *Event, client *Client) {
	switch event.Type {
	case "subscribe":
		go client.Subscribe(relay, event.Channel)
		break
	case "message":
		relay.Publish(event)
		break
	default:
		log.Printf("%s", event)
	}
}

func (relay *Relay) Publish(event *Event) {
	if event.Channel == "" {
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
