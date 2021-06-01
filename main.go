package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"regexp"
	"ungo/game"
)

var hubs = make(map[string]*game.Hub)
var mainLobby *game.Hub

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//Load the environment variables from the projectvars.env file
func initEnv() {
	if _, err := os.Stat("projectvars.env"); err == nil {
		err = godotenv.Load("projectvars.env")
		if err != nil {
			fmt.Println("Error loading environment.env")
		}
		fmt.Println("Current environment:", os.Getenv("ENV"))
	}
}
func main() {
	initEnv()
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
			for _, hub := range hubs{
				delete(hub.Clients, ws)
			}

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
		broadcastPlayerChange(mainLobby)
		defer func() {
			delete(hubs, username)
			keys := make([]string, 0, len(hubs))
			for k := range hubs {
				keys = append(keys, k)
			}
			jsonString, _ := json.Marshal(keys)

			// Broadcast all current lobbies to main lobby
			mainLobby.Broadcast <- game.OutgoingMessage{
				Event:  "NewLobby",
				Message: string(jsonString),
			}
		}()
		keys := make([]string, 0, len(hubs))
		for k := range hubs {
			keys = append(keys, k)
		}
		jsonString, _ := json.Marshal(keys)

		// Broadcast all current lobbies to main lobby
		mainLobby.Broadcast <- game.OutgoingMessage{
			Event:  "NewLobby",
			Message: string(jsonString),
		}
		// Listen on connection
		read(mainLobby, ws, username)
	})
	println(":"+os.Getenv("OUTPORT"))

	err := http.ListenAndServe(":"+os.Getenv("OUTPORT"), r)
	if err != nil {
		log.Println("Unable to bind to port")
		return 
	}
}

func broadcastPlayerChange(lobby *game.Hub) {
	var players []string
	for _, player := range lobby.Clients {
		players = append(players, player)
	}
	playersJSON, _ := json.Marshal(players)
	lobby.Broadcast <- game.OutgoingMessage{
		Event:   "PlayerChange",
		Message: string(playersJSON),
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
			usernameTaken := playerExists(temp)
			if !usernameTaken {
				username = temp
				response.Event = "Registered"
				response.Message = username
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

func playerExists(username string) bool {
	for _,player := range mainLobby.Clients {
		if username == player{
			return true
		}
	}

	for _, hub := range hubs {
		for _,player := range hub.Clients {
			if username == player{
				return true
			}
		}
	}
	return false
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

	// Broadcast all current lobbies to main lobby
	mainLobby.Broadcast <- game.OutgoingMessage{
		Event:  "NewLobby",
		Message: string(jsonString),
	}

	joinLobby(newLobby, client, username, username)
	return true
}

func joinLobby(lobby *game.Hub, client *websocket.Conn, lobbyName string, username string) {
	if _, ok := hubs[lobbyName]; ok {
		// Create a lobby
		lobby = game.NewHub()
		go lobby.Run()

		hubs[lobbyName] = lobby
		lobby.Clients[client] = username
		players := make([]string, 0, len(lobby.Clients))
		for _, player := range lobby.Clients {
			players = append(players, player)
		}
		jsonString, _ := json.Marshal(players)

		lobby.Broadcast <- game.OutgoingMessage{
			Event:  "JoinedLobby",
			Message: username,
		}
		lobby.Broadcast <- game.OutgoingMessage{
			Event:  "PlayerChange",
			Message: string(jsonString),
		}
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
	players := make([]string, 0, len(lobby.Clients))
	for _, player := range lobby.Clients {
		players = append(players, player)
	}
	jsonString, _ := json.Marshal(players)
	broadcast := game.OutgoingMessage{
		Event:   "PlayerChange",
		Message: string(jsonString),
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
