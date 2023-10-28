package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Server struct {
	Address         string
	Upgrader        *websocket.Upgrader
	ClientConnected chan *Client
}

func NewServer(address string) *Server {
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return &Server{
		Address:         address,
		Upgrader:        upgrader,
		ClientConnected: make(chan *Client),
	}
}

func (server *Server) Listen() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", server.Upgrade)

	_server := &http.Server{
		Addr:         server.Address,
		Handler:      mux,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}
	defer _server.Shutdown(context.Background())

	err := _server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func (server *Server) Upgrade(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	if params["token"] == nil {
		log.Println("Connect request made without token")
		return
	}

	conn, err := server.Upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("Error upgrading connection")
		panic(err)
	}

	token := string(params["token"][0])

	// send the new client to the clients connected channel
	server.ClientConnected <- NewClient(conn, token)
}
