import React, { useCallback, useEffect, useState } from 'react';
import './App.css';

const socket = new WebSocket("ws://127.0.0.1:3000/ws");

function CreateLobby() {
  return null;
}

function JoinLobby() {
  return null;
}

function App() {
  const [message, setMessage] = useState('')
  const [inputValue, setInputValue] = useState('')
  const [lobbyName, setLobbyName] = useState('')

  useEffect(() => {
    socket.onopen = () => {
      setMessage('Connected')
    };

    socket.onmessage = (e) => {
      setMessage("Get message from server: " + e.data)
    };

    return () => {
      socket.close()
    }
  }, [])

  const handleClick = useCallback((e) => {
    e.preventDefault()

    socket.send(JSON.stringify({
      action: "",
      message: inputValue
    }))
  }, [inputValue])

  const handleChange = useCallback((e) => {
    setInputValue(e.target.value)
  }, [])

  const handleCreateLobby = useCallback((e) => {
    e.preventDefault()
    socket.send(JSON.stringify({
      action: "CreateLobby",
      message: lobbyName
    }))
  }, [lobbyName])

  const handleLobbyChange = useCallback((e) => {
    setLobbyName(e.target.value)
  }, [])

  return (
      <div className="App">
        <CreateLobby/>
        <JoinLobby/>
        <label>
            Create Lobby
          <input id="input" type="text" value={lobbyName} onChange={handleLobbyChange} />
          <button onClick={handleCreateLobby}>Send</button>
        </label>
        <input id="input" type="text" value={inputValue} onChange={handleChange} />
        <button onClick={handleClick}>Send</button>
        <pre>{message}</pre>
      </div>
  );
}

export default App;