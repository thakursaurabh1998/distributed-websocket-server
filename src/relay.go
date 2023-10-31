package src

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

type Relay struct {
	Server      *Server
	Redis       *Redis
	ClientStore *ClientStore
	Channels    *Channels
}

func (relay *Relay) Start() {
	log.Println("Starting HTTP Server")
	go relay.Server.Listen()

	if relay.Redis != nil {
		relay.subscribeRedisChannels()
	}

	for client := range relay.Server.ClientConnected {
		relay.handleConnection(client)
	}
}

func (relay *Relay) handleConnection(client *Client) {
	relay.ClientStore.Lock()
	relay.ClientStore.Clients[client.ID] = client
	go client.Process(relay)

	log.Println("Clients: ", len(relay.ClientStore.Clients))
	relay.ClientStore.Unlock()
}

func (relay *Relay) HandleIncomingClientEvent(event *Event, client *Client) {
	switch event.Type {
	case "log":
		go relay.handleLog(event, client)
	case "command":
		commandEventData := &CommandEventData{}

		if err := json.Unmarshal(event.Data, commandEventData); err != nil {
			log.Println("JSON decoding error", err)
			return
		}

		go relay.Channels.HandleIncomingEvent(relay, client, event.Channel, commandEventData)
	default:
		log.Printf("Event type not supported: %s", event.Type)
	}
}

func (relay *Relay) handleLog(event *Event, client *Client) {
	log.Printf("Log event received from client: %s", client.ID)
}

func (relay *Relay) PublishRedis(clientId string, event *OutgoingEvent) {
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

func (relay *Relay) subscribeRedisChannels() {
	channels := reflect.ValueOf(*relay.Channels)
	for i := 0; i < channels.NumField(); i += 1 {
		method := channels.Field(i).MethodByName("ChannelName")
		value := method.Call(nil)

		go relay.subscribeRedis(value[0].String())
	}
}

func (relay *Relay) subscribeRedis(channel string) {
	psc := relay.Redis.Client.Subscribe(channel)
	defer psc.Close()

	log.Printf("Subscribed to redis channel: %s", channel)

	for {
		message, err := psc.ReceiveMessage()
		Panic(err)

		event := &OutgoingEvent{}

		json.Unmarshal([]byte(message.Payload), event)
		go relay.handleIncomingRedisEvent(event)
	}
}

func (relay *Relay) handleIncomingRedisEvent(event *OutgoingEvent) {
	if client := relay.ClientStore.Clients[event.ClientID]; client != nil {
		client.Emit(event.Channel, event.Data)
	}
}

func (relay *Relay) PublishMessage(clientId, channel string, payload json.RawMessage) {
	event := &OutgoingEvent{
		Event: &Event{
			Type:    "message",
			Channel: channel,
			Data:    payload,
		},
		ClientID: clientId,
	}

	if client := relay.ClientStore.Clients[clientId]; client != nil {
		client.Emit(channel, payload)
	} else {
		if relay.Redis != nil {
			relay.PublishRedis(clientId, event)
		}
	}
}
