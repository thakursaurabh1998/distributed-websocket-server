package main

import "encoding/json"

type Event struct {
	Type    string
	Channel string
	Data    json.RawMessage
}

