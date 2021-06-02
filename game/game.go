package game

import (
	"math/rand"
	"time"
)

type Card struct{
	Type string
	Number int
	Color string
	WildCardColor string
}

var colors = []string{"green", "yellow", "red", "blue"}
func GenerateCard() Card{
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(100)

	if n < 3 {
		return Card{
			Type:   "Wild",
		}
	}
	if n >= 3 && n < 6 {
		return Card{
			Type:   "Plus4",
		}
	}
	if n >= 6 && n < 9 {
		return Card{
			Type:   "Reverse",
		}
	}
	if n >= 9 && n < 14 {
		return Card{
			Type:   "Skip",
			Number: 0,
			Color:  colors[rand.Intn(4)],
		}
	}
	if n >= 14 && n < 17 {
		return Card{
			Type:   "Plus2",
			Number: 0,
			Color:  colors[rand.Intn(4)],
		}
	}

	return Card{
		Type:   "Number",
		Number: rand.Intn(10),
		Color:  colors[rand.Intn(4)],
	}
}