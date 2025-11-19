package game

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"slices"
	"strconv"
	"sync"
	"yamb/broadcaster"

	"golang.org/x/net/websocket"
)

type Room struct {
	Mu          sync.Mutex
	Broadcaster *broadcaster.Broadcaster

	ID           string
	Players      []*Player
	Dice         *Dice
	CurrentTurn  int // index of the player whose turn it is
	GameStarted  bool
	NumOfPlayers int // 2-4
	NumOfDice    int // 5 or 6

	ChatConns map[*websocket.Conn]bool
}

func NewRoom(mode, dice string) *Room {
	numOfDice, _ := strconv.Atoi(dice)
	numOfPlayers := 2
	switch mode {
	case "1v1":
		numOfPlayers = 2
	case "1v1v1":
		numOfPlayers = 3
	case "2v2":
		numOfPlayers = 4
	}
	return &Room{
		Broadcaster: broadcaster.NewBroadcaster(),

		Players:      []*Player{},
		CurrentTurn:  0,
		GameStarted:  false,
		Dice:         NewDice(numOfDice),
		NumOfPlayers: numOfPlayers,
		NumOfDice:    numOfDice,

		ChatConns: make(map[*websocket.Conn]bool),
	}
}

func (r *Room) IsFull() bool {
    r.Mu.Lock()
    defer r.Mu.Unlock()
    return len(r.Players) == r.NumOfPlayers
}

func (r *Room) AddPlayer(player *Player) error {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.Players = append(r.Players, player)
	r.Players[len(r.Players)-1].Team = Team(len(r.Players) - 1)
	return nil
}

// TODO: move to dice.go
func (r *Room) RollDice() {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	if r.Dice.RollsLeft > 0 {
		for i := range r.NumOfDice {
			if !r.Dice.Held[i] {
				r.Dice.Values[i] = 1 + (rand.Intn(6))
			}
		}
		r.Dice.RollsLeft--
	}
}

func (r *Room) EndTurn() {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.CurrentTurn = (r.CurrentTurn + 1) % len(r.Players)
	r.Dice = NewDice(r.NumOfDice)
}

func (r *Room) GameEnded() bool {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	for _, p := range r.Players {
		if !p.ScoreCard.IsComplete() {
			return false
		}
	}
	return true
}

func (r *Room) GetPlayerByID(playerID string) *Player {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	for _, p := range r.Players {
		if p.ID == playerID {
			return p
		}
	}
	return nil
}

// used after game ends to sort players by score in order to announce winner
func (r *Room) SortPlayersByScore() {
	if !r.GameEnded() {
		return
	}

	r.Mu.Lock()
	defer r.Mu.Unlock()

	sorted := make([]*Player, len(r.Players))
	copy(sorted, r.Players)
	slices.SortFunc(sorted, func(a, b *Player) int {
		return b.ScoreCard.TotalScore() - a.ScoreCard.TotalScore()
	})
	r.Players = sorted
}

func (r *Room) AddConn(ws *websocket.Conn) {
	fmt.Printf("Adding new connection: %v\n", ws.RemoteAddr())

	r.Mu.Lock()
	r.ChatConns[ws] = true
	r.Mu.Unlock()

	r.readLoop(ws)
}

func (r *Room) RemoveConn(ws *websocket.Conn) {
	fmt.Printf("Removing connection: %v\n", ws.RemoteAddr())

	r.Mu.Lock()
	delete(r.ChatConns, ws)
	r.Mu.Unlock()
}

func (r *Room) readLoop(ws *websocket.Conn) {
	buffer := make([]byte, 1024)
	for {
		n, err := ws.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Connection closed by client: %v\n", ws.RemoteAddr())
				r.RemoveConn(ws)
				ws.Close()
				break
			}
			fmt.Printf("Error reading from connection %v: %v\n", ws.RemoteAddr(), err)
			r.RemoveConn(ws)
			ws.Close()
			continue
		}
		msg := buffer[:n]
		fmt.Printf("Received message from %v: %s\n", ws.RemoteAddr(), string(msg))
		ws.Write([]byte("Message received"))
		r.Broadcast(string(msg))
	}
}

func (r *Room) Broadcast(msg string) {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	for ws := range r.ChatConns {
		go func(ws *websocket.Conn) {
			// send as HTML fragment for htmx ws extension
			_, err := ws.Write([]byte(msg))
			if err != nil {
				fmt.Printf("Error broadcasting to %v: %v\n", ws.RemoteAddr(), err)
				r.RemoveConn(ws)
				ws.Close()
			}
		}(ws)
	}
}
