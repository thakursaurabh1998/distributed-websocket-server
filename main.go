package main

import (
	"log"
)

func main() {
	log.Println("Starting relay")

	relay := &Relay{
		ClientStore: NewClientStore(),
		Redis:       NewRedis("localhost:6379"),
		Server:      NewServer("0.0.0.0:8081"),
	}

	done := make(chan bool)

	relay.Start()
	<-done
}
