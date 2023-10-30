package main

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

func (mscCns *messageConsumer) HandleCommand(command string, payload interface{}) {
	messageConsumerPayload := payload.(*MessageConsumerPayload)
	switch command {
	case "ready":
		mscCns.ready(messageConsumerPayload)
	default:
		log.Printf("%s", payload)
	}
}

func (mscCns *messageConsumer) ready(payload *MessageConsumerPayload) {
	log.Println("MessageConsumer ready")
}
