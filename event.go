package main

import "encoding/json"

type Event struct {
	Type    string          `json:"type"`
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}
