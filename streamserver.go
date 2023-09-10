// StreamServer sends payload data over websocket.

package main

import (
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

type StreamServer struct{
	webHandle	chan []byte
	register	chan chan []byte
	unregister	chan chan []byte
	done		chan string
}

func NewStreamServer() *StreamServer {
	return &StreamServer{
		webHandle:	make(chan []byte, 100),
		register:	make(chan chan []byte, 10),
		unregister:	make(chan chan []byte, 10),
		done:		make(chan string, 10),
	}
}

func (s *StreamServer) start() {
	go func() {
		http.Handle("/endpoint", websocket.Handler(s.acceptAndHold))
		if err := http.ListenAndServe(":2222", nil); err != nil {
			log.Panic("StreamServer failed to ListendAndServe", err)
		}
	}()
	log.Println("Started websocket")

	go func() {
		clients := make(map[chan []byte]bool, 100)
		for {
			select {
			case connectedClient := <-s.register:
				log.Println("Registered client")
				clients[connectedClient] = true
			case disconnectedClient := <-s.unregister:
				delete(clients, disconnectedClient)
			case appData := <-s.webHandle:
				for client, _ := range clients {
					client<-appData
				}
			case msg := <-s.done:
				//close client connections
				log.Println("Returning from streamServer.start()", msg)
				return
			}
		}
	}()
	log.Println("Started appdata distributor")
}

func (s *StreamServer) acceptAndHold(ws *websocket.Conn) {
	log.Println("Opening connection")

	dataChannel := make(chan []byte, 100)
	s.register<-dataChannel
	defer func(){ s.unregister<-dataChannel }()

	log.Println("Registering datachannel")

	done := make(chan string, 10)
	go func(){
		var discardableMessage string
		for{
			log.Println("Waiting for message")
			if err := websocket.Message.Receive(ws, &discardableMessage); err != nil {
				done<-"Connection Closed"
				return
			}
		}
	}()

	log.Println("Accepted Connection", ws.LocalAddr())

	websocket.Message.Send(ws, "Hello")

	for {
		select {
		case msg := <-done:
			log.Println("Returning", msg, ws.LocalAddr())
			return
		case appData := <-dataChannel:
			if err := websocket.Message.Send(ws, appData); err != nil {
				log.Println("Failed to send message", ws.LocalAddr())
			}
		}
	}
}
