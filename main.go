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
			if _, ok := mainLobby.Clients[ws] ; ok{
				removeClientFromHub(mainLobby, ws)
				broadcastPlayerChange(mainLobby)
			}
			for _, hub := range hubs{
				if _, ok := hub.Clients[ws] ; ok{
					removeClientFromHub(hub, ws)
					broadcastPlayerChange(hub)
				}
			}

			err := ws.Close()
			if err != nil {
				return
			}
			log.Printf("Closed!")
		}()
		log.Println("Connected!")

		username := registerUser(mainLobby, ws)
		mainLobby.Clients[ws] = game.Player{
			Username: username,
			Hand:     nil,
		}
		broadcastPlayerChange(mainLobby)
		defer func() {
			fmt.Println("Player Disconnected: ", username)
			if hub, ok := hubs[username]; ok {
				endGame(hub)
				return
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
		players = append(players, player.Username)
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
			removeClientFromHub(lobby, client)
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
		if username == player.Username{
			return true
		}
	}

	for _, hub := range hubs {
		for _,player := range hub.Clients {
			if username == player.Username{
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
			removeClientFromHub(lobby, client)
			break
		}
		println("Attempting action: ", message.Action)

		switch message.Action {
		case "CreateLobby":
			lobby = createLobby(lobby, client, username)
			break
		case "DeleteLobby":
			break
		case "JoinLobby":
			lobby = joinLobby(client, message.Message, username)
			break
		case "StartGame":
			startGame(lobby, client, username)
			break
		case "TakeTurn":
			takeTurn(lobby, client, message.TurnInfo, username)
			break
		case "ReturnToMainLobby":
			lobby = returnToMainLobby(lobby, client, username)
			break
		case "SendMessageToLobby":
			sendMessageToLobby(lobby, message.Message)
			break
		}
		println("Completed action: ", message.Action)
	}
}

func startGame(lobby *game.Hub, client *websocket.Conn, username string) {
	lobby.Mu.Lock()
	defer lobby.Mu.Unlock()
	if hubs[username] != lobby{
		if err := client.WriteJSON(game.OutgoingMessage{
			Event:   "PermissionDenied",
		}); !errors.Is(err, nil) {
			log.Printf("error occurred: %v", err)
		}
	}
	lobby.GameStarted = true
	cardCounts := make(map[string]int)
	for conn, player := range lobby.Clients {
		cardCounts[player.Username] = 7
		for i := 0; i < 7; i++ {
			player.Hand = append(player.Hand, game.GenerateCard())
		}
		jsonString,_ := json.Marshal(player.Hand)
		if err := conn.WriteJSON(game.OutgoingMessage{
			Event:   "HandChanged",
			Message: string(jsonString),
		}); !errors.Is(err, nil) {
			log.Printf("error occurred: %v", err)
		}
		lobby.Clients[conn] = player
	}

	lobby.MostRecentCard = game.GenerateCard()
	lobby.CurrentTurn = username

	lobby.Broadcast <- game.OutgoingMessage{
		Event:   "GameStarted",
		Message: username,
		TurnInfo: lobby.MostRecentCard,
		CardCounts: cardCounts,
	}
}

func createLobby(lobby *game.Hub, client *websocket.Conn, username string) *game.Hub{
	removeClientFromHub(lobby, client)
	newLobby := game.NewHub()
	go newLobby.Run()
	lobby.Mu.Lock()
	hubs[username] = newLobby
	lobby.Mu.Unlock()

	player := newLobby.Clients[client]
	player.Username = username
	lobby.Mu.RLock()
	keys := make([]string, 0, len(hubs))
	for k := range hubs {
		keys = append(keys, k)
	}
	lobby.Mu.RUnlock()
	jsonString, _ := json.Marshal(keys)

	// Broadcast all current lobbies to main lobby
	mainLobby.Broadcast <- game.OutgoingMessage{
		Event:  "NewLobby",
		Message: string(jsonString),
	}

	joinLobby(client, username, username)
	return newLobby
}

func joinLobby(client *websocket.Conn, lobbyName string, username string) *game.Hub{
	if lobby, ok := hubs[lobbyName]; ok {
		if len(lobby.Clients) > 4{
			if err := client.WriteJSON(game.OutgoingMessage{
				Event:   "CannotJoin",
			}); !errors.Is(err, nil) {
				log.Printf("error occurred: %v", err)
			}
			return mainLobby
		}
		player := lobby.Clients[client]
		player.Username = username

		if err := client.WriteJSON(game.OutgoingMessage{
			Event:   "JoinedLobby",
			Message: lobbyName,
		}); !errors.Is(err, nil) {
			log.Printf("error occurred: %v", err)
		}
		removeClientFromHub(mainLobby, client)
		lobby.Clients[client] = player

		broadcastPlayerChange(lobby)
		broadcastPlayerChange(mainLobby)
		return lobby
	}
	if err := client.WriteJSON(game.OutgoingMessage{
		Event:   "CannotJoin",
	}); !errors.Is(err, nil) {
		log.Printf("error occurred: %v", err)
	}

	return mainLobby
}

func takeTurn(lobby *game.Hub, client *websocket.Conn, playerCard game.Card, username string) {
	lobby.Mu.RLock()
	if lobby.CurrentTurn == username{
		player := lobby.Clients[client]
		playerHasCard := false
		var cardIndex int
		for curCardIndex, card := range player.Hand {
			if card == playerCard{
				cardIndex = curCardIndex
				playerHasCard = true
			}
		}
		if !playerHasCard{
			if err := client.WriteJSON(game.OutgoingMessage{
				Event:   "CardIsInvalid",
			}); !errors.Is(err, nil) {
				log.Printf("error occurred: %v", err)
			}
			return
		}

		topCard := lobby.MostRecentCard
		cardIsvalid := false
		if topCard.Color == playerCard.Color{
			cardIsvalid = true
		}
		if topCard.Type == "Number" && topCard.Number == playerCard.Number{
			cardIsvalid = true
		}
		if topCard.Type == "Wild" && topCard.Color == playerCard.Color{
			cardIsvalid = true
		}
		if playerCard.Type == "Wild" {
			if playerCard.Color == "green" || playerCard.Color == "yellow" ||
				playerCard.Color == "red" || playerCard.Color == "blue" {
				cardIsvalid = true
			}
		}

		if cardIsvalid{
			topCard = playerCard
			nextPlayer := nextUser(lobby, username)
			lobby.Mu.RUnlock()
			lobby.Mu.Lock()
			switch topCard.Type {
			case "Plus2":
				for i := 0; i < 2; i++ {
					nextPlayer.Hand = append(nextPlayer.Hand, game.GenerateCard())
				}
				break
			case "Plus4":
				for i := 0; i < 4; i++ {
					nextPlayer.Hand = append(nextPlayer.Hand, game.GenerateCard())
				}
				break
			case "Skip":
				nextPlayer = nextUser(lobby, nextPlayer.Username)
				break
			case "Reverse":
				lobby.Clockwise = !lobby.Clockwise
				nextPlayer = nextUser(lobby, username)

				break
			}
			cardCounts := make(map[string]int)
			for _, p := range lobby.Clients {
				cardCounts[p.Username] = len(p.Hand)
			}
			lobby.CurrentTurn = nextPlayer.Username
			lobby.MostRecentCard = game.GenerateCard()
			lobby.Broadcast <- game.OutgoingMessage{
				Event:   "NextTurn",
				Message: lobby.CurrentTurn,
				TurnInfo: lobby.MostRecentCard,
				CardCounts: cardCounts,
			}
			lobby.Mu.Unlock()
		}else{
			if err := client.WriteJSON(game.OutgoingMessage{
				Event:   "CardIsInvalid",
			}); !errors.Is(err, nil) {
				log.Printf("error occurred: %v", err)
			}
		}

		player.Hand = append(player.Hand[:cardIndex], player.Hand[cardIndex+1:]...)

	} else {
		if err := client.WriteJSON(game.OutgoingMessage{
			Event:   "TurnUnavailable",
		}); !errors.Is(err, nil) {
			log.Printf("error occurred: %v", err)
		}
	}
}

func nextUser(lobby *game.Hub, username string) *game.Player {
	lobby.Mu.Lock()
	defer lobby.Mu.Unlock()
	var p *game.Player
	if len(lobby.Clients) == 1 {
		endGame(lobby)
	}
	if lobby.Clockwise{
		nextIsTurn := false

		for _, player := range lobby.Clients {
			if nextIsTurn {
				nextIsTurn = false
				p = &player
				break
			}
			if player.Username == username {
				nextIsTurn = true
			}
		}
		if nextIsTurn == true {
			for _, player := range lobby.Clients {
				p = &player
				break
			}
		}
	}else{
		i := 0
		var temp *game.Player
		for _, player := range lobby.Clients {
			temp = &player
			if player.Username == username{
				if i != 0 {
					break
				}
			}
			i++
		}

		p = temp
	}
	return p
}

func returnToMainLobby(lobby *game.Hub, client *websocket.Conn, username string) *game.Hub{
	removeClientFromHub(lobby, client)
	mainLobby.Clients[client] = game.Player{
		Username: username,
		Hand:     nil,
	}
	player := mainLobby.Clients[client]
	player.Username = username
	// Return all current lobbies
	if err := client.WriteJSON(game.OutgoingMessage{
		Event:   "ReturnedToMainLobby",
	}); !errors.Is(err, nil) {
		log.Printf("error occurred: %v", err)
	}
	if hubs[username] == lobby{
		endGame(lobby)
 	}

	broadcastPlayerChange(mainLobby)
	return mainLobby
}

func endGame(lobby *game.Hub) {
	for con, player := range lobby.Clients {
		returnToMainLobby(lobby, con, player.Username)
	}
	var keys []string
	for s, hub := range hubs {
		if hub == lobby{
			delete(hubs, s)
		}else{
			keys = append(keys, s)
		}
	}
	jsonString, _ := json.Marshal(keys)

	// Broadcast all current lobbies to main lobby
	mainLobby.Broadcast <- game.OutgoingMessage{
		Event:  "LobbyChange",
		Message: string(jsonString),
	}
}

func removeClientFromHub(hub *game.Hub, client *websocket.Conn){
	hub.Mu.Lock()
	defer hub.Mu.Unlock()
	delete(hub.Clients, client)
}

func sendMessageToLobby(lobby *game.Hub, message string) {
	broadcast := game.OutgoingMessage{
		Event:   "NewMessage",
		Message: message,
	}

	lobby.Broadcast <- broadcast
}