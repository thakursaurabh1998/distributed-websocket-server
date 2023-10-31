package src

import "encoding/json"

type Event struct {
	// Event types can be "log", "command", "config"
	Type string `json:"type"`

	// Channel types are same as the features
	// eg: "announcements", "messages" or "x"
	Channel string `json:"channel"`

	// Data is the payload of the event
	Data json.RawMessage `json:"data"`
}

type OutgoingEvent struct {
	*Event

	// ClientID is the ID of the client to
	// which the event needs to be transmitted
	ClientID string `json:"client_id"`
}

type CommandEventData struct {
	Command string          `json:"command"`
	Payload json.RawMessage `json:"payload"`
}
