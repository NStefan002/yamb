package game

// enum for team
type Team int

const (
	Blue Team = iota
	Red
	Yellow
)

type Player struct {
	ID        string
	Username  string
	ScoreCard ScoreCard
	Team      Team
	// TODO: add connection for chat
}

func NewPlayer(id, username string) *Player {
	return &Player{
		ID:        id,
		Username:  username,
		ScoreCard: NewScoreCard(),
	}
}
