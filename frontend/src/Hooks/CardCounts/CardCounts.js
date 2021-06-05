import React from 'react';

export default function CardCounts(props) {
    if (props.hands.length === 0) {
        return null
    }
    return (
        <div>
            <h3>Card Counts</h3>
            {Object.keys(props.hands).map((player, k) => (
                <div key={k}>{player} {props.hands[player]}</div>
            ))}
        </div>
    );
}