import React from 'react';

export default function Lobbies(props) {
    if(props.lobbyList.length === 0 || props.lobbyList){
        return null
    }
    let lobbiesArray = JSON.parse(props.lobbyList)
    const joinLobby = (lobbyName) => {
        props.ws.send(JSON.stringify({
            action: "JoinLobby",
            card: lobbyName
        }))
    }

    return (
        <div>
            Lobbies
            {lobbiesArray != null && lobbiesArray.map((lobby)=>(
                <button onClick={()=>joinLobby(lobby)}>{lobby}'s game</button>
            ))}
        </div>
    );

}