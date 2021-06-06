# Compile and Play
To compile by hand, you will first need to install Go (Language) and NPM+NodeJS 
## Compiling Go code
`Go run build main.go` will compile the code along with install of the dependencies

## Compiling the React Frontend
First navigate the frontend folder
Then run ``npm install``
And finally run ``npm start``

## Playing the game
You should now be able to play the game from `localhost:3000`

Step 1) Connect to `Localhost:3000`

Step 2) Enter a unique username. The game will warn you if the name is taken and as you to try again

Step 3) Join a game listed under lobbies or create your own game

Step 4) If you are the host, start the game when another user joins or simply wait for the host to start the game

Step 5) The rules of uno are simple, play a card matching the top cards color or type. Special cards must match types (Skip/Plus 2/ Plus 4)
and number cards must match their number values. Or you can choose to draw a new card if you do not have a matching card to play.

Step 6) Get down to 0 cards

Step 7) Start the game again!

# How it works
## Connecting to the server
Websockets are used to maintain connections between the server and the client.
Websockets are message oriented rather than byte-oriented. 
The biggest difference between the two is perhaps the need to ensure that messages are not being send at the same time per channel.
Either side of the connection can send and receive a message at the same time but neither side can send or receive multiple messages at the same time.
This creates the need for mutex locking the connection. Behind the scenes a websocket is byte-oriented since it uses a TCP connection.
However, each message is treated as a message rather than a byte-stream from the perspective of the developer.

## Connection Hubs
A hub is a collection of user connections along with their usernames. 
This information is stored as a map with the connection being the key.

## Lobbies
Lobbies are simply hubs that also contain game information. 
The Main Lobby is the only lobby that is a pure connection hub as it does not contain any game information.
Game lobbies contain all information that hubs do along with game information such as current hands, turns, and if the game is currently in session.

## Hub hub
A map of all game lobbies is stored in a map called `Hubs` that uses the username of the user who created the lobby as the key.

## Outgoing messages
All outgoing messages follow the following format:
`` {
Event string `json:"event"`
Message string `json:"message"`
TurnInfo Card `json:"card_payload"`
CardCounts map[string]int `json:"card_count"`
}``
The message contains information about the event ie if a card was played, the next turn, a new message in the lobby etc.
The "Message" portion contains the primary contents of the message and the TurnInfo and CardCounts are sent along to ensure the user is always up to date.
This method is midly inefficient as information is passed on every message rather than only when something changes.

## Incoming messages
All Incoming messages follow the following format:
``{
Action string `json:"action"`
Message string `json:"message"`
TurnInfo Card `json:"card_payload"`
}``

The action contains information about the action a user is taking such as playing a card or drawing a new one.
The message portion contains information not related to their turn such as the message they want to broadcast to 
other plays or the username they would like to register.
TurnInfo contains the card the player wants to play.

## Register
To register a client, a websocket connection must first be established at `HOSTNAME:PORT/ws`
Once the connection is established, no message will be accepted unless the Action field is set to `Register` and the `Message` field contains a unique username
At which point the users connection will be added to the MainLobby connection hub and all other users in the hub will be notified.

## List Games
Games are automatically listed and updated to any user in the MainLobby connection hub. 
Users inside a game Hub will not be notified of any new lobbies or any lobbies being deleted.

## Create Game
To create a game the `Action` field must be set to `CreateLobby` with all other information being ignored if there is any.
The game will be added to the global `Hubs` map with the name of the lobby being the username of the user who created the lobby. 
The user who created the hub will be automatically moved from the MainLobby to the new lobby.
 
## Join Game
To join a game the `Action` field must be set to `JoinLobby` and the `Message` must be set to the name of the lobby. 
If the lobby does not exist or the `Message` is empty, then the user will receeive a message notifying them the lobby isn't real.
**The user can only join a hub from the MainLobby and not while in another game lobby**

## Exit Game
To exit a game the `Action` field must be set to `ReturnToMainLobby`
The user will then be moved from their hub and into the MainLobby.
If the user is the host, then the game will automatically be ended and the lobby will be deleted. 
**The user can exit a game if they are in a game and not in the MainLobby**

## Unregister
To unregister, a user must simply close the tab at which point they will be removed from any lobby they are in and any
lobbies they host will be deleted. Users will be notified of any changes to players in the lobby.

## Broadcast a message (Chat Feature)
To send a message the `Action` feautre must be set to `SendMessageToLobby` and the `Message` field set to the message to be broadcast
