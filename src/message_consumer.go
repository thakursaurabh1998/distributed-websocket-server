package src

import (
	"log"
)

type MessageConsumerPayload struct {
	VisitorId string `json:"visitor_id"`
}

// MessageConsumer is a consumer that handles events from the message_consumer channel
type messageConsumer struct{}

func NewMessageConsumer() Channel {
	return &messageConsumer{}
}

func (mscCns *messageConsumer) ChannelName() string {
	return "message_consumer"
}

func (mscCns *messageConsumer) HandleCommand(relay *Relay, client *Client, command string, payload interface{}) {
	messageConsumerPayload := payload.(*MessageConsumerPayload)
	switch command {
	case "ready":
		mscCns.ready(relay, client, messageConsumerPayload)
	default:
		log.Printf("%s", payload)
	}
}

func (mscCns *messageConsumer) ready(relay *Relay, client *Client, payload *MessageConsumerPayload) {
	log.Println("MessageConsumer ready")
	data := []byte(`{"hello": "world"}`)
	relay.PublishMessage(client.ID, mscCns.ChannelName(), MESSAGE, data)
}
