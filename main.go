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
var players = make(map[string]int)

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
		log.Println("Connected!")

		username := registerUser(mainLobby, ws)
		// Add client
		mainLobby.Clients[ws] = username
		println("Players:")
		for player := range players {
			println(player)
		}
		// Listen on connection
		read(mainLobby, ws, username)
	})
	err := http.ListenAndServe(":3000", r)
	if err != nil {
		log.Println("Unable to bind to port")
		return 
	}
}

func registerUser(lobby *game.Hub, client *websocket.Conn) string {
	username := ""
	for username == "" {
		var message game.IncomingMessage
		err := client.ReadJSON(&message)
		if !errors.Is(err, nil) {
			log.Printf("error occurred: %v", err)
			delete(lobby.Clients, client)
			break
		}

		if message.Action == "Register" {
			reg, err := regexp.Compile("[^a-zA-Z0-9]+")
			if err != nil {
				log.Fatal(err)
			}
			temp := reg.ReplaceAllString(message.Message, "")
			var response game.OutgoingMessage
			if _, ok := players[temp]; !ok {
				username = temp
				response.Event = "Registered"
				response.Message = username
				players[username] = 0
			}else{
				response.Event = "UsernameInUse"
			}
			if err := client.WriteJSON(response); !errors.Is(err, nil) {
				log.Printf("error occurred: %v", err)
			}
		}
	}

	return username
}

func read(lobby *game.Hub, client *websocket.Conn, username string) {
	for {
		var message game.IncomingMessage
		err := client.ReadJSON(&message)
		if !errors.Is(err, nil) {
			log.Printf("error occurred reading message: %v", err)
			delete(lobby.Clients, client)
			break
		}
		switch message.Action {
		case "CreateLobby":
			createLobby(lobby, client, username)
			break
		case "DeleteLobby":
			break
		case "JoinLobby":
			joinLobby(lobby, client, message.Message, username)
			break
		case "ReturnToMainLobby":
			returnToMainLobby(lobby, client, username)
			break
		case "SendMessageToLobby":
			sendMessageToLobby(lobby, message.Message)
			break
		}
		println("Completed action: ", message.Action)
	}
}

func createLobby(lobby *game.Hub, client *websocket.Conn, username string) bool{
	delete(lobby.Clients, client)
	newLobby := game.NewHub()
	go newLobby.Run()

	hubs[username] = newLobby
	newLobby.Clients[client] = username

	keys := make([]string, 0, len(hubs))
	for k := range hubs {
		keys = append(keys, k)
	}
	jsonString, _ := json.Marshal(keys)

	mainLobbyMessage := game.OutgoingMessage{
		Event:  "NewLobby",
		Message: string(jsonString),
	}

	// Broadcast all current lobbies to main lobby
	mainLobby.Broadcast <- mainLobbyMessage

	newLobbyBroadcast := game.OutgoingMessage{
		Event:  "PlayerChange",
		Message: string(jsonString),
	}
	players := make([]string, 0, len(newLobby.Clients))

	for _, player := range newLobby.Clients {
		players = append(players, player)
	}
	jsonString, _ = json.Marshal(players)
	newLobbyBroadcast.Message = string(jsonString)
	//Broadcast all players to current lobby
	newLobby.Broadcast <- newLobbyBroadcast
	return true
}

func joinLobby(lobby *game.Hub, client *websocket.Conn, lobbyName string, username string) {
	if _, ok := hubs[lobbyName]; !ok {
		// Create a lobby
		lobby = game.NewHub()
		go lobby.Run()

		hubs[lobbyName] = lobby
		lobby.Clients[client] = username
		message := game.OutgoingMessage{
			Event:  "PlayerChange",
			Message: lobbyName,
		}
		lobby.Broadcast <- message
	}
}

func returnToMainLobby(lobby *game.Hub, client *websocket.Conn, username string) {
	delete(lobby.Clients, client)
	lobby = mainLobby
	lobby.Clients[client] = username
	// Return all current lobbies
	if err := client.WriteJSON(game.OutgoingMessage{
		Event:   "ReturnedToMainLobby",
	}); !errors.Is(err, nil) {
		log.Printf("error occurred: %v", err)
	}
	broadcast := game.OutgoingMessage{
		Event:   "PlayerChange",
	}
	lobby.Broadcast <- broadcast
}

func sendMessageToLobby(lobby *game.Hub, message string) {
	broadcast := game.OutgoingMessage{
		Event:   "NewMessage",
		Message: message,
	}

	lobby.Broadcast <- broadcast
}
