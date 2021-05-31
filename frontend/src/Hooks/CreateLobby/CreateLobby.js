import React from 'react';

export default function CreateLobby(props) {
    const handleSubmit = (e) => {
        e.preventDefault()
        props.ws.send(JSON.stringify({
            action: "CreateLobby",
            message: ""
        }))
    }

    return (
        <label>
            <button onClick={(e) => handleSubmit(e)}>
                Create Lobby
            </button>
        </label>
    );
}