import {React, useEffect} from 'react';

export default function Chat(props) {
    useEffect(() => {
    }, [props.messages])

    return (
        <div>
            <h3>Chat</h3>
            {props.messages.map((message, k)=>(
                <div key={k}>{message}</div>
            ))}
        </div>
    );
}