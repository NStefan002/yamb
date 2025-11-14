package game

import (
	"errors"
	"slices"
)

type Dice struct {
	Values    []int
	Held      []bool
	RollsLeft int
}

func NewDice(num int) *Dice {
	values := make([]int, num)
	held := make([]bool, num)
	for i := range num {
		values[i] = i + 1
		held[i] = false
	}
	return &Dice{
		Values:    values,
		Held:      held,
		RollsLeft: 3,
	}
}

func (d *Dice) ToggleDie(index int) {
	if index >= 0 && index < len(d.Held) {
		d.Held[index] = !d.Held[index]
	}
}

func (d *Dice) getHeldDice() []int {
	held := []int{}
	for i, v := range d.Values {
		if d.Held[i] {
			held = append(held, v)
		}
	}
	return held
}

func (d *Dice) counts() map[int]int {
	held := d.getHeldDice()
	counts := make(map[int]int)
	for _, v := range held {
		counts[v]++
	}
	return counts
}

func (d *Dice) sum() int {
	held := d.getHeldDice()
	sum := 0
	for _, v := range held {
		sum += v
	}
	return sum
}

func (d *Dice) Number(value int) (int, error) {
	held := d.getHeldDice()
	sum := 0
	for _, v := range held {
		if v == value {
			sum += v
		} else {
			return 0, errors.New("not all dice match the value")
		}
	}
	return sum, nil
}

func (d *Dice) MinMax() (int, error) {
	held := d.getHeldDice()
	if len(held) != 5 {
		return 0, errors.New("need all 5 dice")
	}
	return d.sum(), nil
}

func (d *Dice) Kenta() (int, error) {
	held := d.getHeldDice()
	if len(held) != 5 {
		return 0, errors.New("need all 5 dice")
	}
	sorted := make([]int, len(held))
	copy(sorted, held)
	slices.Sort(sorted)

	// check for small kenta (1-5)
	smallKenta := true
	for i := range 5 {
		if sorted[i] != i+1 {
			smallKenta = false
			break
		}
	}
	if smallKenta {
		return 55, nil
	}

	// check for big kenta (2-6)
	bigKenta := true
	for i := range 5 {
		if sorted[i] != i+2 {
			bigKenta = false
			break
		}
	}
	if bigKenta {
		return 60, nil
	}

	// no kenta (no error because we want to allow user to write 0)
	return 0, nil
}

func (d *Dice) Full() (int, error) {
	held := d.getHeldDice()
	if len(held) != 5 {
		return 0, errors.New("need all 5 dice")
	}

	counts := d.counts()
	hasThree := false
	hasTwo := false

	for _, count := range counts {
		if count == 3 {
			hasThree = true
			continue
		}
		if count == 2 {
			hasTwo = true
		}
	}
	if !hasThree || !hasTwo {
		// no full (no error because we want to allow user to write 0)
		return 0, nil
	}
	// 2 same + 3 same + 30
	return d.sum() + 30, nil
}

func (d *Dice) Poker() (int, error) {
	held := d.getHeldDice()
	if len(held) != 4 {
		return 0, errors.New("need 4 dice")
	}
	for i := range len(held) - 1 {
		// all should be the same
		if held[i] != held[i+1] {
			// no poker (no error because we want to allow user to write 0)
			return 0, nil
		}
	}
	// 4 same + 50
	return held[0]*4 + 50, nil
}

func (d *Dice) Yamb() (int, error) {
	held := d.getHeldDice()
	if len(held) != 5 {
		return 0, errors.New("need all 5 dice")
	}
	for i := range len(held) - 1 {
		// all should be the same
		if held[i] != held[i+1] {
			// no yamb (no error because we want to allow user to write 0)
			return 0, nil
		}
	}
	// 5 same + 80
	return held[0]*5 + 80, nil
}
