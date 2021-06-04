package game

import (
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type IncomingMessage struct {
	Action string `json:"action"`
	Message string `json:"message"`
	TurnInfo Card `json:"card_payload"`
}

type OutgoingMessage struct {
	Event string `json:"event"`
	Message string `json:"message"`
	TurnInfo Card `json:"card_payload"`
	CardCounts map[string]int `json:"card_count"`
}

type Hub struct{
	Clients   map[*websocket.Conn]Player
	TurnOrder []*websocket.Conn
	Broadcast chan OutgoingMessage
	GameStarted bool
	Clockwise bool
	CurrentTurn string
	MostRecentCard Card
	Mu sync.RWMutex
}

type Player struct{
	Username string
	Hand    []Card
}

func NewHub() *Hub {
	return &Hub{
		Clients:   make(map[*websocket.Conn]Player),
		Broadcast: make(chan OutgoingMessage),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case message := <-h.Broadcast:
			h.Mu.Lock()
			for client, _ := range h.Clients {
				if err := client.WriteJSON(message); !errors.Is(err, nil) {
					log.Printf("error occurred: %v", err)
				}
			}
			h.Mu.Unlock()
		}
	}
}
