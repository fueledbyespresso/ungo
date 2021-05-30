package game

import (
	"errors"
	"github.com/gorilla/websocket"
	"log"
)

type Message struct {
	Action string `json:"action"`
	Message string `json:"message"`
}

type Hub struct{
	Clients   map[*websocket.Conn]bool
	Broadcast chan Message
	PlayerListChange chan []string
}

func NewHub() *Hub {
	return &Hub{
		Clients:   make(map[*websocket.Conn]bool),
		Broadcast: make(chan Message),
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