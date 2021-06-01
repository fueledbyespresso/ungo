import React, {useEffect, useState} from 'react';
import CreateLobby from "./Hooks/CreateLobby/CreateLobby";
import JoinLobby from "./Hooks/JoinLobby/JoinLobby";
import Players from "./Hooks/Players/Players";
import Chat from "./Hooks/Chat/Chat";
import Register from "./Hooks/Register/Register";
import Lobbies from "./Hooks/Lobbies/Lobbies";

const socket = new WebSocket("ws://127.0.0.1:3000/ws");

function App() {
    const [username, setUserName] = useState(null)
    const [inMainLobby, setInMainLobby] = useState(false)
    const [messages, setMessages] = useState([])
    const [inputValue, setInputValue] = useState('')
    const [lobbies, setLobbies] = useState([])
    const [lobbyName, setLobbyName] = useState('')
    const [playerList, setPlayerList] = useState([])

    useEffect(() => {
        socket.onopen = () => {
            console.log("Connected")
        };

        socket.onmessage = (e) => {
            let data = JSON.parse(e.data)
            setMessages(oldMessages => [...oldMessages, data.event])

            switch (data.event) {
                case "NewMessage":
                    setMessages(oldMessages => [...oldMessages, data.event])
                    break
                case "Registered":
                    setInMainLobby(true)
                    setUserName(data.message)
                    break
                case "PlayerChange":
                    setPlayerList(JSON.parse(data.message))
                    break
                case "ReturnedToMainLobby":
                    setInMainLobby(true)
                    break
                case "JoinedLobby":
                    setLobbyName(data.message)
                    setInMainLobby(false)
                    break
                case "NewLobby":
                    setLobbies(data.message)
                    break
                default:
                    break
            }
        };
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
            action: "ReturnedToMainLobby",
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
            <div className="main-lobby">
                <h2>Main Lobby</h2>
                <CreateLobby ws={socket}/>
                <Lobbies lobbyList={lobbies} ws={socket}/>
                <Chat messages={messages}/>
                <Players playerList={playerList}/>
            </div>
        )
    }else {
        return (
            <div className="App">
                <h2>{lobbyName}'s Game</h2>
                <button onClick={()=>returnToMainLobby}>Return to Main Menu</button>
                <Players playerList={playerList}/>
                <Chat messages={messages}/>
            </div>
        );
    }
}

export default App;