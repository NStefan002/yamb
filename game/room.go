package game

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
			{ID: "1", Username: "Alice", ScoreCard: NewScoreCard(), Dice: NewDice()},
			{ID: "2", Username: "Bob", ScoreCard: NewScoreCard(), Dice: NewDice()},
			{ID: "3", Username: "Charlie", ScoreCard: NewScoreCard(), Dice: NewDice()},
		},
		CurrentTurn: 0,
		GameStarted: false,
	}
}

func (g *Room) AddPlayer(player *Player) {
	g.Players = append(g.Players, player)
}
