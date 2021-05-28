package main

import (
	"errors"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"ungo/game"
)

var hubs map[string]*game.Hub

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	r := gin.Default()

	r.Use(static.Serve("/", static.LocalFile("./frontend/build", true)))

	// Create a hub
	lobby := game.NewHub()

	// Start a go routine
	go lobby.Run()
	r.GET("/ws", func(c *gin.Context) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if !errors.Is(err, nil) {
			log.Println(err)
		}
		defer func() {
			delete(lobby.Clients, ws)
			err := ws.Close()
			if err != nil {
				return 
			}
			log.Printf("Closed!")
		}()

		// Add client
		lobby.Clients[ws] = true

		log.Println("Connected!")

		// Listen on connection
		read(lobby, ws)
	})
	err := http.ListenAndServe(":3000", r)
	if err != nil {
		log.Println("Unable to bind to port")
		return 
	}
}

func read(lobby *game.Hub, client *websocket.Conn) {
	for {
		var message game.Message
		err := client.ReadJSON(&message)
		if !errors.Is(err, nil) {
			log.Printf("error occurred: %v", err)
			delete(lobby.Clients, client)
			break
		}
		if message.Action == "CreateLobby"{
			createLobby(message.Message, client)
			delete(lobby.Clients, client)
		}
		log.Println(message)

		// Send a message to hub
		lobby.Broadcast <- message
	}
}

func createLobby(lobbyName string, client *websocket.Conn){
	if _, ok := hubs[lobbyName]; ok {
		// Create a lobby
		newLobby := game.NewHub()
		hubs[lobbyName] = newLobby
		newLobby.Clients[client] = true
	}
}