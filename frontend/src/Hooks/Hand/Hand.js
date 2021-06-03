import {React, useEffect} from 'react';
import "./Hand.scss"
import UnoCard from "./UnoCard";

export default function Hand(props) {
    console.log(props.cards)

    if(props.cards.length === 0){
        return null
    }
    const drawCard = () => {
        props.ws.send(JSON.stringify({
            action: "Draw",
        }))
    }
    return (
        <div>
            <h3>Chat</h3>
            <div className={"hand"}>
                {props.cards.map((card, k)=>(
                    <UnoCard cardInfo={card} ws={props.ws} key={k+card.Color+card.Type+card.Number}/>
                ))}
                <button onClick={()=>drawCard()}>Draw Card</button>
            </div>
        </div>
    );
}