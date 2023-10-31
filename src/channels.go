package src

import (
	"encoding/json"
	"log"
)

type Channel interface {
	ChannelName() string
	HandleCommand(relay *Relay, client *Client, command string, payload interface{})
}

type Channels struct {
	MessageConsumer Channel
}

func InitChannels() *Channels {
	return &Channels{
		MessageConsumer: NewMessageConsumer(),
	}
}

func (channels *Channels) HandleIncomingEvent(relay *Relay, client *Client, ftChannel string, eventData *CommandEventData) {
	switch ftChannel {
	case "message_consumer":
		msgcnsPayload := &MessageConsumerPayload{}
		if err := json.Unmarshal(eventData.Payload, msgcnsPayload); err != nil {
			log.Println("JSON decoding error", err)
			return
		}

		channels.MessageConsumer.HandleCommand(relay, client, eventData.Command, msgcnsPayload)
	default:
		log.Printf("Command not available for channel %s: %s", ftChannel, eventData.Command)
	}
}
