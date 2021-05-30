import {React, useState} from 'react';

export default function Register(props) {
    const [input, setInput] = useState(null)

    return (
        <label>
            Enter a username:
            <input value={input} onChange={setInput}/>
            <button onClick={(e)=>props.handleSubmit(e, input)}>
                Enter
            </button>
        </label>
    );

}