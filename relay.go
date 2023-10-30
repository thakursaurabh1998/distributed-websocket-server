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
	Channels     *Channels
}

func (relay *Relay) Start() {
	log.Println("Starting HTTP Server")
	go relay.Server.Listen()

	for client := range relay.Server.ClientConnected {
		relay.HandleConnection(client)
	}
}

func (relay *Relay) HandleConnection(client *Client) {
	relay.ClientStore.Lock()
	relay.ClientStore.Clients[client.ID] = client
	go client.Process(relay)

	log.Println("Clients: ", len(relay.ClientStore.Clients))
	relay.ClientStore.Unlock()
}

func (relay *Relay) HandleIncomingClientEvent(event *Event, client *Client) {
	switch event.Type {
	case "log":
		go relay.HandleLog(event, client)
	case "command":
		commandEventData := &CommandEventData{}

		if err := json.Unmarshal(event.Data, commandEventData); err != nil {
			log.Println("JSON decoding error", err)
			return
		}

		go relay.Channels.HandleIncomingEvent(event.Channel, commandEventData)
	default:
		log.Printf("Event type not supported: %s", event.Type)
	}
}

func (relay *Relay) HandleLog(event *Event, client *Client) {
	log.Printf("Log event received from client: %s", client.ID)
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
