package game

import (
	"errors"
	"github.com/gorilla/websocket"
	"log"
)

type IncomingMessage struct {
	Action string `json:"action"`
	Message string `json:"message"`
}

type OutgoingMessage struct {
	Event string `json:"event"`
	Message string `json:"message"`
}

type Hub struct{
	Clients   map[*websocket.Conn]string
	Broadcast chan OutgoingMessage
	PlayerListChange chan []string
}

func NewHub() *Hub {
	return &Hub{
		Clients:   make(map[*websocket.Conn]string),
		Broadcast: make(chan OutgoingMessage),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case message := <-h.Broadcast:
			for client := range h.Clients {
				if err := client.WriteJSON(message); !errors.Is(err, nil) {
					log.Printf("error occurred: %v", err)
				}
			}
		case message := <-h.Broadcast:
			for client := range h.Clients {
				if err := client.WriteJSON(message); !errors.Is(err, nil) {
					log.Printf("error occurred: %v", err)
				}
			}
		}
	}
}