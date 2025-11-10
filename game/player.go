package game

// enum for team
type Team int

const (
	Blue Team = iota
	Red
	Yellow
)

type Player struct {
	ID         string
	Username   string
	ScoreCard  ScoreCard
	Team       Team
	FinalScore int
}

func NewPlayer(id, username string) *Player {
	return &Player{
		ID:         id,
		Username:   username,
		ScoreCard:  NewScoreCard(),
		FinalScore: 0,
	}
}
