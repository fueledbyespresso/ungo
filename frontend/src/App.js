import React, {useEffect, useState} from 'react';
import CreateLobby from "./Hooks/CreateLobby/CreateLobby";
import Players from "./Hooks/Players/Players";
import Chat from "./Hooks/Chat/Chat";
import Register from "./Hooks/Register/Register";
import Lobbies from "./Hooks/Lobbies/Lobbies";
import Hand from "./Hooks/Hand/Hand";
import CardCounts from "./Hooks/CardCounts/CardCounts";

const HOST = process.env.REACT_APP_HOST
const PORT = process.env.REACT_APP_PORT

let wsPath
if (process.env.NODE_ENV === "production") {
    wsPath = "wss://" + HOST + "/ws"
} else {
    wsPath = "ws://" + HOST + ":" + PORT + "/ws"
}

const socket = new WebSocket(wsPath)

function App() {
    const [isConnected, setIsConnected] = useState(false)
    const [username, setUserName] = useState(null)
    const [inMainLobby, setInMainLobby] = useState(false)
    const [gameStarted, setGameStarted] = useState(false)
    const [messages, setMessages] = useState([])
    const [lobbies, setLobbies] = useState([])
    const [lobbyName, setLobbyName] = useState('')
    const [playerList, setPlayerList] = useState([])
    const [hand, setHand] = useState([])
    const [hands, setHands] = useState([])
    const [currentCard, setCurrentCard] = useState(null)
    const [currentPlayer, setCurrentPlayer] = useState(null)

    useEffect(() => {
        socket.onopen = () => {
            setIsConnected(true)
            console.log("Connected")
        };

        socket.onmessage = (e) => {
            let data = JSON.parse(e.data)

            switch (data.event) {
                case "NewMessage":
                    setMessages(oldMessages => [...oldMessages,
                        {
                            "type": "user",
                            "sender": data.message.sender,
                            "message": data.message.message
                        }])
                    break
                case "Registered":
                    setInMainLobby(true)
                    setUserName(data.message)
                    setMessages(oldMessages => [...oldMessages,
                        {
                            "type": "SYSTEM",
                            "message": "Welcome, " + data.message
                        }])
                    break
                case "PlayerChange":
                    setPlayerList(JSON.parse(data.message))
                    break
                case "ReturnedToMainLobby":
                    setInMainLobby(true)
                    setMessages(oldMessages => [...oldMessages,
                        {
                            type: "SYSTEM",
                            message: "Returning to main lobby"
                        }])
                    break
                case "JoinedLobby":
                    setLobbyName(data.message)
                    setInMainLobby(false)
                    setMessages(oldMessages => [...oldMessages,
                        {
                            type: "SYSTEM",
                            message: "Joined " +data.message +"'s lobby"
                        }])
                    break
                case "NewLobby":
                    setLobbies(data.message)
                    break
                case "LobbyChange":
                    setLobbies(data.message)
                    break
                case "GameStarted":
                    setCurrentCard(data.card_payload)
                    setCurrentPlayer(data.message)
                    setHands(data.card_count)
                    setGameStarted(true)

                    break
                case "HandChanged":
                    setHand(JSON.parse(data.message))
                    break
                case "NextTurn":
                    setCurrentPlayer(data.message)
                    setHands(data.card_count)
                    setCurrentCard(data.card_payload)
                    break
                default:
                    break
            }
        };

        socket.onclose = () => {
            setIsConnected(false)
            setMessages(oldMessages => [...oldMessages,
                {
                    type: "SYSTEM",
                    message: "You have been disconnected from the server"
                }])
            console.log("Connection closed")
        }
    }, [lobbies, messages])

    const handleSubmit = (e, username) => {
        e.preventDefault()
        socket.send(JSON.stringify({
            action: "Register",
            message: username
        }))
    }

    const returnToMainLobby = () => {
        socket.send(JSON.stringify({
            action: "ReturnToMainLobby",
        }))
    }

    const startGame = () => {
        socket.send(JSON.stringify({
            action: "StartGame",
        }))
    }
    // Force user to register on initial load
    if (username == null) {
        return (
            <div className="App">
                <Register handleSubmit={handleSubmit}/>
            </div>
        )
    }

    if (inMainLobby) {
        return (
            <div className="app">
                <h2>Main Lobby</h2>
                <CreateLobby ws={socket}/>
                <Lobbies lobbyList={lobbies} ws={socket}/>
                <Players playerList={playerList}/>
                <Chat messages={messages}/>
            </div>
        )
    } else {
        return (
            <div className="App">
                <div>Hello, {username}</div>
                <h2>{lobbyName}'s Game</h2>
                <button onClick={() => returnToMainLobby()}>Return to Main Menu</button>
                <Players playerList={playerList}/>
                <CardCounts hands={hands}/>
                {username === lobbyName && !gameStarted &&
                (<button onClick={startGame}>Start game</button>)
                }
                {gameStarted &&
                <div>
                    <h3>Top Card: </h3>
                    <div className={"uno-card"} style={{borderColor: currentCard.Color}}>
                        <div>{currentCard.Type}</div>
                        <div>{currentCard.Number}</div>
                        <div>{currentCard.Color}</div>
                    </div>
                </div>
                }
                {gameStarted && <h2>Next Turn: {currentPlayer}</h2>}

                {gameStarted && <Hand cards={hand} ws={socket}/>}
                <Chat messages={messages}/>
            </div>
        );
    }
}

export default App;