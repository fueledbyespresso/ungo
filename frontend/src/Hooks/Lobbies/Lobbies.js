import React from 'react';

export default function Lobbies(props) {
    if(props.lobbyList.length === 0){
        return null
    }
    let lobbiesArray = JSON.parse(props.lobbyList)
    const joinLobby = (lobbyName) => {
        props.ws.send(JSON.stringify({
            action: "JoinLobby",
            message: lobbyName
        }))
    }
    return (
        <div>
            Lobbies
            {lobbiesArray.map((lobby)=>(
                <button onClick={()=>joinLobby(lobby)}>{lobby}'s game</button>
            ))}
        </div>
    );

}