package main

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"regexp"
	"ungo/game"
)

var hubs = make(map[string]*game.Hub)
var mainLobby *game.Hub

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	r := gin.Default()

	r.Use(static.Serve("/", static.LocalFile("./frontend/build", true)))

	// Create a hub
	mainLobby = game.NewHub()

	// Start a go routine
	go mainLobby.Run()
	r.GET("/ws", func(c *gin.Context) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if !errors.Is(err, nil) {
			log.Println(err)
		}
		defer func() {
			delete(mainLobby.Clients, ws)
			err := ws.Close()
			if err != nil {
				return 
			}
			log.Printf("Closed!")
		}()

		// Add client
		mainLobby.Clients[ws] = true

		log.Println("Connected!")

		// Listen on connection
		read(mainLobby, ws)
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
		switch message.Action {
		case "CreateLobby":
			createLobby(lobby, client, message.Message)
		case "DeleteLobby":
		case "JoinLobby":
			joinLobby(lobby, client, message.Message)
		case "ReturnToMainLobby":
			delete(lobby.Clients, client)
			lobby = mainLobby
			lobby.Clients[client] = true
			// Return all current lobbies
			lobby.Broadcast <- message
		case "SendMessage":
			message.Action = "NewMessage"
			lobby.Broadcast <- message
		}
	}
}

func createLobby(lobby *game.Hub, client *websocket.Conn, lobbyName string) bool{
	delete(lobby.Clients, client)
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	processedString := reg.ReplaceAllString(lobbyName, "")
	if _, ok := hubs[processedString]; !ok {
		// Create a lobby
		lobby = game.NewHub()
		go lobby.Run()

		hubs[processedString] = lobby
		lobby.Clients[client] = true
	}
	keys := make([]string, 0, len(hubs))

	for k := range hubs {
		keys = append(keys, k)
	}
	jsonString, _ := json.Marshal(keys)

	message := game.Message{
		Action:  "NewLobby",
		Message: string(jsonString),
	}
	// Broadcast all current lobbies to main lobby
	mainLobby.Broadcast <- message

	players := make([]bool, 0, len(lobby.Clients))

	for _, player := range lobby.Clients {
		players = append(players, player)
	}
	jsonString, _ = json.Marshal(players)
	message.Message = string(jsonString)
	//Broadcast all players to current lobby
	lobby.Broadcast <- message
	return true
}

func joinLobby(lobby *game.Hub, client *websocket.Conn, lobbyName string) {
	if _, ok := hubs[lobbyName]; !ok {
		// Create a lobby
		lobby = game.NewHub()
		go lobby.Run()

		hubs[lobbyName] = lobby
		lobby.Clients[client] = true
		message := game.Message{
			Action:  "PlayerChange",
			Message: lobbyName,
		}
		lobby.Broadcast <- message
	}
}