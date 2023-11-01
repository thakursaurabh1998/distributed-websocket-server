package src

import (
	"encoding/json"
	"fmt"
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

func (relay *Relay) HandleIncomingClientEvent(event *IncomingEvent, client *Client) {
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

func (relay *Relay) handleLog(event *IncomingEvent, client *Client) {
	log.Printf("Log event received from client: %s", client.ID)
}

func (relay *Relay) dispatchEventToRedis(event *OutgoingEvent) {
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
		go relay.dispatchEventToClient(event)
	}
}

func (relay *Relay) dispatchEventToClient(event *OutgoingEvent) {
	switch event.Type {
	case MESSAGE:
		if client := relay.ClientStore.Clients[event.ClientID]; client != nil {
			event.ClientID = ""
			client.Emit(event.Channel, event.Data)
		}
	case BROADCAST:
		for _, client := range relay.ClientStore.Clients {
			client.Emit(event.Channel, event.Data)
		}
	default:
		log.Printf("Event type not supported: %s", event.Type)
	}
}

func (relay *Relay) PublishMessage(clientId, channel string, eventType OutgoingEventType, payload json.RawMessage) error {
	event := &OutgoingEvent{
		Event: &Event{
			Channel: channel,
			Data:    payload,
		},
		Type: eventType,
	}

	if eventType == MESSAGE {
		if clientId == "" {
			return fmt.Errorf("Client ID required for MESSAGE event type")
		}
		event.ClientID = clientId
	}

	if relay.Redis != nil {
		relay.dispatchEventToRedis(event)
	} else {
		relay.dispatchEventToClient(event)
	}

	return nil
}
