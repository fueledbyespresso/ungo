import {React, useState} from 'react';
import "./Register.scss"

export default function Register(props) {
    const [input, setInput] = useState("")

    return (
        <label className="Register">
            Enter a username:
            <input value={input} onChange={e => setInput(e.target.value)}/>
            <button onClick={(e) => props.handleSubmit(e, input)}>
                Enter
            </button>
        </label>
    );
}