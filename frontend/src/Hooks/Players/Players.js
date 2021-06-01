import React from 'react';

export default function Players(props) {
    if(props.playerList.length === 0){
        return null
    }
    return (
        <div>
            <h3>Players</h3>
            {props.playerList.map((player, k)=>(
                <div key={k}>{player}</div>
            ))}
        </div>
    );
}