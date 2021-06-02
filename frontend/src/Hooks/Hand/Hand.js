import {React, useEffect} from 'react';

export default function Hand(props) {
    console.log(props.cards)

    if(props.cards.length === 0){
        return null
    }

    return (
        <div>
            <h3>Chat</h3>
            {props.cards.map((card, k)=>(
                <div key={k}>
                    <div>{card.Type}</div>
                    <div>{card.Number}</div>
                    <div>{card.Color}</div>
                </div>
            ))}
        </div>
    );
}