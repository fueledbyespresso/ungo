import React, {Fragment, useCallback, useEffect, useState} from 'react';
import CreateLobby from "./Hooks/CreateLobby/CreateLobby";
import JoinLobby from "./Hooks/JoinLobby/JoinLobby";
import Players from "./Hooks/Players/Players";
import Chat from "./Chat/Chat";

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
      console.log(data)
      setMessage("Get message from server: " + data.message)
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
        {username != null && (
            <Fragment>
              <CreateLobby/>
              <JoinLobby/>
              <Players/>
              <Chat message = {message}/>

              <label>
                Create Lobby
                <input id="input" type="text" value={lobbyName} onChange={handleLobbyChange} />
                <button onClick={handleCreateLobby}>Send</button>
              </label>
              <input id="input" type="text" value={inputValue} onChange={handleChange} />
              <button onClick={handleClick}>Send</button>
            </Fragment>
        )}
      </div>
  );
}

export default App;