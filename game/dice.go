package game

import (
	"math/rand"
)

type Dice struct {
	Values    []int
	Held      []bool
	RollsLeft int
}

func NewDice() Dice {
	return Dice{
		Values:    []int{1, 2, 3, 4, 5, 6},
		Held:      []bool{false, false, false, false, false, false},
		RollsLeft: 3,
	}
}

func (g *Room) RollDice() {
	// Simple dice rolling logic
	if g.Players[g.CurrentTurn].Dice.RollsLeft > 0 {
		for i := range 6 {
			if !g.Players[g.CurrentTurn].Dice.Held[i] {
				g.Players[g.CurrentTurn].Dice.Values[i] = 1 + (rand.Intn(6))
			}
		}
		g.Players[g.CurrentTurn].Dice.RollsLeft--
	} else {
		// no rolls left, move to next player
		g.CurrentTurn = (g.CurrentTurn + 1) % len(g.Players)
	}
}

func (d *Dice) ToggleDie(index int) {
	if index >= 0 && index < len(d.Held) {
		d.Held[index] = !d.Held[index]
	}
}
