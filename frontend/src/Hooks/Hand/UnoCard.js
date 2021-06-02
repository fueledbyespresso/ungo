import React, {Fragment, useState} from 'react';

export default function UnoCard(props) {
    const [card, setCard] = useState(props.cardInfo)

    const playCard = () => {
        props.ws.send(JSON.stringify({
            action: "TakeTurn",
            card_payload: card,
        }))
    }
    const updateColor = (color) => {
        let copyOfCard = card
        copyOfCard.WildCardColor = color
        copyOfCard.Color = color
        setCard(copyOfCard)
    }
    return (
        <div className={"uno-card"} style={{borderColor: card.Color}}>
            <div>{card.Type}</div>
            {card.Type === "Wild" ? (
                <div>
                    Choose color
                    <button onChange={() => updateColor("green")}>green</button>
                    <button onChange={() => updateColor("yellow")}>yellow</button>
                    <button onChange={() => updateColor("red")}>red</button>
                    <button onChange={() => updateColor("blue")}>blue</button>
                </div>
            ):(
                <Fragment>
                    <div>{card.Number}</div>
                    <div>{card.Color}</div>
                </Fragment>
            )}

            <button onClick={()=>playCard(card)}>Play Card</button>
        </div>
    );

}