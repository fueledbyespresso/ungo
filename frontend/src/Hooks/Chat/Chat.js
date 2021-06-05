import {React, useEffect, useState} from 'react';
import "./Chat.scss"

export default function Chat(props) {
    const [input, setInput] = useState("")

    useEffect(() => {
        console.log(props.messages.length > 0)
    }, [props.messages])

    const sendMessage = () => {
        props.ws.send(JSON.stringify({
            action: "SendMessageToLobby",
            message: input
        }))
        setInput("")
    }

    return (
        <div className="chat">
            <h3>Chat</h3>
            {props.messages.length > 0 && props.messages.map((message, k) => (
                message.type === "user" ? (
                    <div key={k}>{message.sender}: {message.message}</div>
                ):(
                    <div key={k}>{message.message}</div>
                )
            ))}
            <textarea value={input} onChange={e=>setInput(e.target.value)}/>
            <button onClick={()=>sendMessage()}>Send</button>
        </div>
    );
}