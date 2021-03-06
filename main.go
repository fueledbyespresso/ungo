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
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"
	"ungo/game"
)

var hubs = make(map[string]*game.Hub)
var mainLobby *game.Hub
var hubsMU = sync.RWMutex{}

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
	rand.Seed(time.Now().UnixNano())

	// Start a go routine
	go mainLobby.Run()
	r.GET("/ws", func(c *gin.Context) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if !errors.Is(err, nil) {
			log.Println(err)
		}
		defer func() {
			fmt.Println("Removing player from lobbies")
			if _, ok := mainLobby.Clients[ws] ; ok{
				removeClientFromHub(mainLobby, ws)
				broadcastPlayerChange(mainLobby)
			}
			for host, hub := range hubs{
				if player, ok := hub.Clients[ws] ; ok{
					if host == player.Username{
						endGame(hub)
					}
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
		mainLobby.Mu.Lock()
		mainLobby.Clients[ws] = game.Player{
			Username: username,
			Hand:     nil,
		}
		mainLobby.Mu.Unlock()
		broadcastPlayerChange(mainLobby)
		hubsMU.RLock()
		keys := make([]string, 0, len(hubs))
		for k := range hubs {
			keys = append(keys, k)
		}
		jsonString, _ := json.Marshal(keys)
		hubsMU.RUnlock()

		// Broadcast all current lobbies to main lobby
		mainLobby.Broadcast <- game.OutgoingMessage{
			Event:  "NewLobby",
			Message: string(jsonString),
		}
		// Listen on connection
		read(mainLobby, ws, username)
	})

	var port string
	if os.Getenv("ENV") == "PROD" {
		port = os.Getenv("PORT")
	}else{
		port = os.Getenv("DEVPORT")
	}
	println("Running on :" + port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Println("Unable to bind to port")
		return 
	}
}

func broadcastPlayerChange(lobby *game.Hub) {
	var players []string
	lobby.Mu.RLock()
	for _, player := range lobby.Clients {
		players = append(players, player.Username)
	}
	lobby.Mu.RUnlock()

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
	mainLobby.Mu.RLock()
	for _,player := range mainLobby.Clients {
		if username == player.Username{
			return true
		}
	}
	mainLobby.Mu.RUnlock()

	hubsMU.RLock()
	for _, hub := range hubs {
		hub.Mu.RLock()
		for _,player := range hub.Clients {
			if username == player.Username{
				return true
			}
		}
		hub.Mu.RUnlock()
	}
	hubsMU.RUnlock()

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
			if mainLobby != lobby{
				break
			}

			lobby = createLobby(client, username)
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
		case "Draw":
			drawCard(lobby, client, username)
			break
		case "ReturnToMainLobby":
			lobby = returnToMainLobby(lobby, client, username)
			break
		case "SendMessageToLobby":
			sendMessageToLobby(lobby, message.Message, username)
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

	if len(lobby.Clients) == 1 {
		if err := client.WriteJSON(game.OutgoingMessage{
			Event:   "CannotStartGameWith1Player",
		}); !errors.Is(err, nil) {
			log.Printf("error occurred: %v", err)
		}
	}
	lobby.GameStarted = true
	lobby.Clockwise = true
	cardCounts := make(map[string]int)
	for conn, player := range lobby.Clients {
		lobby.TurnOrder = append(lobby.TurnOrder, conn)
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
	colors := []string{"green", "yellow", "red", "blue"}
	lobby.MostRecentCard = game.Card{
		Type:   "Number",
		Number: rand.Intn(10),
		Color:  colors[rand.Intn(4)],
	}
	lobby.CurrentTurn = username

	lobby.Broadcast <- game.OutgoingMessage{
		Event:   "GameStarted",
		Message: username,
		TurnInfo: lobby.MostRecentCard,
		CardCounts: cardCounts,
	}
}

func createLobby(client *websocket.Conn, username string) *game.Hub{
	newLobby := game.NewHub()
	go newLobby.Run()
	hubsMU.Lock()
	hubs[username] = newLobby
	hubsMU.Unlock()

	player := newLobby.Clients[client]
	player.Username = username
	hubsMU.RLock()
	keys := make([]string, 0, len(hubs))
	for k := range hubs {
		keys = append(keys, k)
	}
	hubsMU.RUnlock()
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
	hubsMU.RLock()
	lobby, ok := hubs[lobbyName]
	hubsMU.RUnlock()

	if ok {
		hubsMU.RLock()
		if len(lobby.Clients) > 4 {
			if err := client.WriteJSON(game.OutgoingMessage{
				Event:   "CannotJoin",
			}); !errors.Is(err, nil) {
				log.Printf("error occurred: %v", err)
			}
			return mainLobby
		}
		player := mainLobby.Clients[client]
		hubsMU.RUnlock()

		removeClientFromHub(mainLobby, client)
		hubsMU.RLock()
		lobby.Clients[client] = player
		hubsMU.RUnlock()

		broadcastPlayerChange(lobby)
		broadcastPlayerChange(mainLobby)

		lobby.Mu.Lock()
		if err := client.WriteJSON(game.OutgoingMessage{
			Event:   "JoinedLobby",
			Message: lobbyName,
		}); !errors.Is(err, nil) {
			log.Printf("error occurred: %v", err)
		}
		lobby.Mu.Unlock()

		return lobby
	}

	if err := client.WriteJSON(game.OutgoingMessage{
		Event:   "CannotJoin",
	}); !errors.Is(err, nil) {
		log.Printf("error occurred: %v", err)
	}

	return mainLobby
}

func drawCard(lobby *game.Hub, client *websocket.Conn, username string) {
	lobby.Mu.RLock()
	curTurn := lobby.CurrentTurn
	lobby.Mu.RUnlock()
	if curTurn == username {
		lobby.Mu.Lock()
		player := lobby.Clients[client]
		player.Hand = append(player.Hand, game.GenerateCard())
		lobby.Clients[client] = player
		jsonString,_ := json.Marshal(lobby.Clients[client].Hand)
		if err := client.WriteJSON(game.OutgoingMessage{
			Event:   "HandChanged",
			Message: string(jsonString),
		}); !errors.Is(err, nil) {
			log.Printf("error occurred: %v", err)
		}
		lobby.Mu.Unlock()
	}
}

func takeTurn(lobby *game.Hub, client *websocket.Conn, playerCard game.Card, username string) {
	lobby.Mu.RLock()
	curTurn := lobby.CurrentTurn
	gameStarted := lobby.GameStarted
	lobby.Mu.RUnlock()
	if curTurn == username && gameStarted{
		lobby.Mu.RLock()
		player := lobby.Clients[client]
		lobby.Mu.RUnlock()

		playerHasCard := false
		var cardIndex int
		for curCardIndex, card := range player.Hand {
			if card == playerCard{
				cardIndex = curCardIndex
				playerHasCard = true
				break
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

		lobby.Mu.RLock()
		topCard := lobby.MostRecentCard
		lobby.Mu.RUnlock()

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
			fmt.Println("Card is valid")
			nextPlayer := nextUser(lobby, username)
			fmt.Println("Next player: ", nextPlayer.Username)
			switch playerCard.Type {
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
				lobby.Mu.Lock()
				lobby.Clockwise = !lobby.Clockwise
				lobby.Mu.Unlock()
				nextPlayer = nextUser(lobby, username)
				break
			}
			player.Hand = append(player.Hand[:cardIndex], player.Hand[cardIndex+1:]...)

			if len(player.Hand) == 0 {
				lobby.Broadcast <- game.OutgoingMessage{
					Event:   "GameWon",
					Message: player.Username,
				}
				lobby.Mu.Lock()
				lobby.GameStarted = false
				lobby.Mu.Unlock()

				return
			}else{
				lobby.Mu.RLock()
				jsonString,_ := json.Marshal(lobby.Clients[client].Hand)
				lobby.Mu.RUnlock()

				if err := client.WriteJSON(game.OutgoingMessage{
					Event:   "HandChanged",
					Message: string(jsonString),
				}); !errors.Is(err, nil) {
					log.Printf("error occurred: %v", err)
				}
			}

			lobby.Mu.Lock()
			lobby.Clients[client] = player
			lobby.Mu.Unlock()

			lobby.Mu.RLock()
			cardCounts := make(map[string]int)
			for _, p := range lobby.Clients {
				cardCounts[p.Username] = len(p.Hand)
			}
			lobby.Mu.RUnlock()

			lobby.CurrentTurn = nextPlayer.Username
			lobby.MostRecentCard = playerCard
			lobby.Broadcast <- game.OutgoingMessage{
				Event:   "NextTurn",
				Message: lobby.CurrentTurn,
				TurnInfo: lobby.MostRecentCard,
				CardCounts: cardCounts,
			}
		}else{
			if err := client.WriteJSON(game.OutgoingMessage{
				Event:   "CardIsInvalid",
			}); !errors.Is(err, nil) {
				log.Printf("error occurred: %v", err)
			}
		}

	} else {
		if err := client.WriteJSON(game.OutgoingMessage{
			Event:   "TurnUnavailable",
		}); !errors.Is(err, nil) {
			log.Printf("error occurred: %v", err)
		}
	}
}

func nextUser(lobby *game.Hub, username string) game.Player {
	lobby.Mu.RLock()
	defer lobby.Mu.RUnlock()
	var playerPos int
	if len(lobby.Clients) == 1 {
		fmt.Println("Ending game")
		endGame(lobby)
	}
	if lobby.Clockwise{
		for index, playerConnection := range lobby.TurnOrder {
			fmt.Println("---",index, " : ", lobby.Clients[playerConnection].Username)
			if username == lobby.Clients[playerConnection].Username{
				playerPos = index + 1
				break
			}
		}

		if playerPos == len(lobby.TurnOrder){
			playerPos = 0
		}
	}else{
		for index, playerConnection := range lobby.TurnOrder {
			fmt.Println("---",index, " : ", lobby.Clients[playerConnection].Username)
			if username == lobby.Clients[playerConnection].Username{
				playerPos = index - 1
				break
			}
		}

		if playerPos == -1{
			playerPos = len(lobby.TurnOrder) - 1
		}
	}
	fmt.Println("Position of next player: ", playerPos)

	return lobby.Clients[lobby.TurnOrder[playerPos]]
}

func returnToMainLobby(lobby *game.Hub, client *websocket.Conn, username string) *game.Hub{
	removeClientFromHub(lobby, client)
	broadcastPlayerChange(lobby)

	mainLobby.Clients[client] = game.Player{
		Username: username,
		Hand:     nil,
	}

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
	lobby.Mu.RLock()
	for con, player := range lobby.Clients {
		lobby.Mu.RUnlock()
		returnToMainLobby(lobby, con, player.Username)
		lobby.Mu.RLock()
	}
	lobby.Mu.RUnlock()

	var keys []string
	hubsMU.Lock()
	for s, hub := range hubs {
		if hub == lobby{
			delete(hubs, s)
		}else{
			keys = append(keys, s)
		}
	}
	hubsMU.Unlock()
	jsonString, _ := json.Marshal(keys)
	// Broadcast all current lobbies to main lobby
	mainLobby.Broadcast <- game.OutgoingMessage{
		Event:  "LobbyChange",
		Message: string(jsonString),
	}
}

func removeClientFromHub(hub *game.Hub, client *websocket.Conn){
	hub.Mu.RLock()
	username := hub.Clients[client].Username
	currentTurn := hub.CurrentTurn
	isMainLobby := hub == mainLobby
	count := len(hub.TurnOrder)
	hub.Mu.RUnlock()

	if (username == currentTurn || !isMainLobby) && count > 1{
		fmt.Println("finding next user")
		hub.CurrentTurn = nextUser(hub, username).Username
	}
	hub.Mu.Lock()
	delete(hub.Clients, client)
	for index, playConnection := range hub.TurnOrder {
		if playConnection == client{
			hub.TurnOrder = append(hub.TurnOrder[:index], hub.TurnOrder[index+1:]...)
		}
	}
	hub.Mu.Unlock()
}

func sendMessageToLobby(lobby *game.Hub, message string, username string) {
	messagePayload := struct {
		Sender string `json:"sender"`
		Message string `json:"message"`
	} {
		username,
		message,
	}

	messageString, err := json.Marshal(messagePayload)
	if err != nil {
		return
	}
	broadcast := game.OutgoingMessage{
		Event:   "NewMessage",
		Message: string(messageString),
	}

	lobby.Broadcast <- broadcast
}