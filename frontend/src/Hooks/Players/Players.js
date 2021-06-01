import React from 'react';

export default function Players(props) {
    console.log(props.playerList)
    if(props.playerList.length === 0){
        return null
    }
    return (
        <div>
            <h3>Players</h3>
            {props.playerList.map((player)=>(
                <div>{player}</div>
            ))}
        </div>
    );
}