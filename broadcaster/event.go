package broadcaster

type EventName string

const (
	PlayerJoined    EventName = "playerJoined"
	ScoreUpdated    EventName = "scoreUpdated"
	DiceAreaUpdated EventName = "diceAreaUpdated"
	CellSelected    EventName = "cellSelected"
	TurnEnded       EventName = "turnEnded"
	ScoreAnnounced  EventName = "scoreAnnounced"
	GameEnded       EventName = "gameEnded"
)

type Event struct {
	Name EventName
}
