package game

type Player struct {
	ID       string
	Username  string
	ScoreCard ScoreCard
	// TODO: add connection for chat
}

func NewPlayer(id, username string) *Player {
	return &Player{
		ID:       id,
		Username:  username,
		ScoreCard: NewScoreCard(),
	}
}
