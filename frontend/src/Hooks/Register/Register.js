import {React, useState} from 'react';

export default function Register(props) {
    const [input, setInput] = useState("")

    return (
        <label>
            Enter a username:
            <input value={input} onChange={e => setInput(e.target.value)}/>
            <button onClick={(e) => props.handleSubmit(e, input)}>
                Enter
            </button>
        </label>
    );
}