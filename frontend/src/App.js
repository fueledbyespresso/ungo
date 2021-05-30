import React, {Fragment, useCallback, useEffect, useState} from 'react';
import CreateLobby from "./Hooks/CreateLobby/CreateLobby";
import JoinLobby from "./Hooks/JoinLobby/JoinLobby";
import Players from "./Hooks/Players/Players";
import Chat from "./Hooks/Chat/Chat";
import Register from "./Hooks/Register/Register";

const socket = new WebSocket("ws://127.0.0.1:3000/ws");

function App() {
    const [username, setUserName] = useState(null)

    const [message, setMessage] = useState('')
    const [inputValue, setInputValue] = useState('')
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
                setMessage(data.message)
                break;
              case "Registered":
                setUserName(data.message)
                break;
              default:
                break
            }
            setMessage("Get message from server: " + data.message)
        };

        return () => {
            socket.close()
        }
    }, [])

    const handleSubmit = (e, username) => {
      e.preventDefault()
      socket.send(JSON.stringify({
        action: "Register",
        message: username
      }))
    }
    // Force user to register on initial load
    if(username == null){
      return (
          <div className="App">
            <Register handleSubmit={handleSubmit}/>
          </div>
      )
    }

    return (
        <div className="App">
            <CreateLobby/>
            <JoinLobby/>
            <Players/>
            <Chat message={message}/>
        </div>
    );
}

export default App;