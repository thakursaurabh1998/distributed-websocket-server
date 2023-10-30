package main

import (
	"flag"
	"log"
)

func main() {
	useRedis := flag.Bool("redis", false, "Use Redis for distributed message pub/sub")
	flag.Parse()

	// Log to stdout with file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Starting relay")
	log.Println("Using Redis:", *useRedis)

	var redis *Redis
	if *useRedis {
		redis = NewRedis("localhost:6379")
	}

	relay := &Relay{
		ClientStore: NewClientStore(),
		Redis:       redis,
		Server:      NewServer("0.0.0.0:8081"),
		Channels:    InitChannels(),
	}

	done := make(chan bool)

	relay.Start()
	<-done
}
