package game

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

func (d *Dice) ToggleDie(index int) {
	if index >= 0 && index < len(d.Held) {
		d.Held[index] = !d.Held[index]
	}
}
