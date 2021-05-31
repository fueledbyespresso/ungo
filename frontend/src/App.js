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
    const [message, setMessage] = useState([])
    const [inputValue, setInputValue] = useState('')
    const [lobbies, setLobbies] = useState('')
    const [lobbyName, setLobbyName] = useState('')
    const [playerList, setPlayerList] = useState('')

    useEffect(() => {
        socket.onopen = () => {
            setMessage('Connected')
        };

        socket.onmessage = (e) => {
            let data = JSON.parse(e.data)
            switch (data.event) {
                case "NewMessage":
                    setMessage(oldMessages => [...oldMessages, data.message]);
                    break
                case "Registered":
                    setUserName(data.message)
                    setInMainLobby(true)
                    break
                case "PlayerChange":
                    setPlayerList(data.message)
                    break
                case "ReturnedToMainLobby":
                    setInMainLobby(true)
                    break
                case "NewLobby":
                    console.log(data.message)
                    setLobbies(data.message)
                    break
                default:
                    break
            }
            setMessage("Get message from server: " + data.message)
        };
    }, [lobbies])

    const handleSubmit = (e, username) => {
        e.preventDefault()
        socket.send(JSON.stringify({
            action: "Register",
            message: username
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
              <Lobbies lobbyList={lobbies}/>
            </div>
        )
    }
    return (
        <div className="App">
            <JoinLobby/>
            <Players/>
            <Chat message={message}/>
        </div>
    );
}

export default App;