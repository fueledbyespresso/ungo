import {React, useEffect} from 'react';
import "./Chat.scss"

export default function Chat(props) {
    useEffect(() => {
        console.log(props.messages.length > 0)
    }, [props.messages])

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
            <textarea/>
            <button>Send</button>
        </div>
    );
}