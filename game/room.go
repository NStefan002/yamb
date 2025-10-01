package game

import "math/rand"

type Room struct {
	ID          string
	Players     []*Player
	CurrentTurn int  // index of the player whose turn it is
	GameStarted bool // true if the game has started
}

func NewRoom() *Room {
	return &Room{
		Players: []*Player{
			// dummy players for testing
			// {ID: "1", Username: "Alice", ScoreCard: NewScoreCard(), Dice: NewDice()},
			// {ID: "2", Username: "Bob", ScoreCard: NewScoreCard(), Dice: NewDice()},
			// {ID: "3", Username: "Charlie", ScoreCard: NewScoreCard(), Dice: NewDice()},
		},
		CurrentTurn: 0,
		GameStarted: false,
	}
}

func (r *Room) AddPlayer(player *Player) {
	r.Players = append(r.Players, player)
}

func (r *Room) RollDice() {
	currPlayer := r.Players[r.CurrentTurn]
	if currPlayer.Dice.RollsLeft > 0 {
		for i := range 6 {
			if !currPlayer.Dice.Held[i] {
				currPlayer.Dice.Values[i] = 1 + (rand.Intn(6))
			}
		}
		currPlayer.Dice.RollsLeft--
	} else {
		r.EndTurn()
	}
}

func (r *Room) EndTurn() {
	r.CurrentTurn = (r.CurrentTurn + 1) % len(r.Players)
	currPlayer := r.Players[r.CurrentTurn]
	currPlayer.Dice = NewDice()
}

func (r *Room) GetPlayerByID(playerID string) *Player {
	for _, p := range r.Players {
		if p.ID == playerID {
			return p
		}
	}
	return nil
}
