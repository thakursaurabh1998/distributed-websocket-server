package main

import (
	"flag"
	"log"

	"github.com/thakursaurabh1998/distributed-websocket-server/src"
)

func main() {
	useRedis := flag.Bool("redis", false, "Use Redis for distributed message pub/sub")
	flag.Parse()

	// Log to stdout with file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Starting relay")
	log.Println("Using Redis:", *useRedis)

	var redis *src.Redis
	if *useRedis {
		redis = src.NewRedis("localhost:6379")
	}

	relay := &src.Relay{
		ClientStore: src.NewClientStore(),
		Redis:       redis,
		Server:      src.NewServer("0.0.0.0:8081"),
		Channels:    src.InitChannels(),
	}

	done := make(chan bool)

	relay.Start()
	<-done
}
