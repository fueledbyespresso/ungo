import React from 'react';

export default function Lobbies(props) {
    if(props.lobbyList.length === 0){
        return null
    }
    let lobbiesArray = JSON.parse(props.lobbyList)
    return (
        <div>
            {lobbiesArray.map((lobby)=>(
                <div>{lobby}</div>
            ))}
        </div>
    );

}