package src

import "encoding/json"

type Event struct {
	// Channel types are same as the features
	// eg: "announcements", "messages" or "x"
	Channel string `json:"channel"`

	// Data is the payload of the event
	Data json.RawMessage `json:"data"`
}

type OutgoingEventType string
type IncomingEventType string

const (
	MESSAGE   OutgoingEventType = "message"
	BROADCAST OutgoingEventType = "broadcast"
)

const (
	LOG     IncomingEventType = "log"
	COMMAND IncomingEventType = "command"
)

type OutgoingEvent struct {
	*Event

	// Event types can be "message", "broadcast"
	Type OutgoingEventType `json:"type"`
	// ClientID is the ID of the client to
	// which the event needs to be transmitted
	ClientID string `json:"client_id"`
}

type IncomingEvent struct {
	*Event

	// Event types can be "log", "command", "config"
	Type IncomingEventType `json:"type"`
}

type CommandEventData struct {
	Command string          `json:"command"`
	Payload json.RawMessage `json:"payload"`
}
