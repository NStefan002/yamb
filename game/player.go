package game

type Player struct {
	ID       string
	Username  string
	ScoreCard ScoreCard
	Dice      Dice
	// TODO: add connection for chat
}

func NewPlayer(id, username string) *Player {
	return &Player{
		ID:       id,
		Username:  username,
		ScoreCard: NewScoreCard(),
		Dice:      NewDice(),
	}
}

func (p *Player) ResetForNewTurn() {
	p.Dice = NewDice()
}
