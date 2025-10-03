package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"yamb/broadcaster"
	"yamb/game"
	"yamb/views"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var (
	rooms   = make(map[string]*game.Room)
	roomsMu sync.Mutex
)

func CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		log.Println("error parsing form:", err)
		return
	}

	roomID := fmt.Sprintf("%06d", rand.Intn(1000000))
	mode := r.FormValue("mode")
	dice := r.FormValue("dice")

	roomsMu.Lock()
	rooms[roomID] = game.NewRoom(mode, dice)
	roomsMu.Unlock()

	err := views.RoomLink(roomID).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "could not render room link", http.StatusInternalServerError)
		log.Println("error rendering room link:", err)
		return
	}
}

func RoomLinkHandler(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	roomsMu.Lock()
	_, ok := rooms[roomID]
	roomsMu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	err := views.UsernameEntry(roomID).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "could not render username entry", http.StatusInternalServerError)
		log.Println("error rendering username entry:", err)
		return
	}
}

func JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.FormValue("room_id")
	username := r.FormValue("username")

	roomsMu.Lock()
	room, ok := rooms[roomID]
	roomsMu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	if len(room.Players) >= 4 {
		http.Error(w, "room full", http.StatusForbidden)
		return
	}

	playerID := uuid.New().String()
	http.SetCookie(w, &http.Cookie{
		Name:  "player_id",
		Value: playerID,
		Path:  "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "room_id",
		Value: roomID,
		Path:  "/",
	})

	err := room.AddPlayer(game.NewPlayer(playerID, username))
	if err != nil {
		http.Error(w, fmt.Sprintf("could not add player: %v", err), http.StatusInternalServerError)
		log.Println("error adding player to room:", err)
		return
	}

	room.Broadcaster.Broadcast(broadcaster.Event{Name: "playerJoined"})
	room.Broadcaster.Broadcast(broadcaster.Event{Name: "scoreUpdated"})

	if len(room.Players) == room.NumOfPlayers {
		room.GameStarted = true
	}

	http.Redirect(w, r, fmt.Sprintf("/room/%s", roomID), http.StatusSeeOther)
}

func RoomPageHandler(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	roomsMu.Lock()
	room, ok := rooms[roomID]
	roomsMu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	playerCookie, err := r.Cookie("player_id")
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/room/%s", roomID), http.StatusSeeOther)
		log.Println("no player cookie:", err)
		return
	}

	playerID := playerCookie.Value
	err = views.RoomPage(roomID, playerID, room).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "could not render room page", http.StatusInternalServerError)
		log.Println("error rendering room page:", err)
		return
	}
}

func RollDiceHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.FormValue("room_id")
	roomsMu.Lock()
	room, ok := rooms[roomID]
	roomsMu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	room.RollDice()
	room.Broadcaster.Broadcast(broadcaster.Event{Name: "diceAreaUpdated"})
	room.Broadcaster.Broadcast(broadcaster.Event{Name: "scoreUpdated"})

	playerCookie, err := r.Cookie("player_id")
	if err != nil {
		http.Error(w, "no player cookie", http.StatusForbidden)
		log.Println("no player cookie:", err)
		return
	}
	playerID := playerCookie.Value

	err = views.DiceArea(roomID, playerID, room).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "could not render dice area", http.StatusInternalServerError)
		log.Println("error rendering dice area:", err)
		return
	}
}

func ToggleDiceHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.FormValue("room_id")
	roomsMu.Lock()
	room, ok := rooms[roomID]
	roomsMu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	dieIdx, _ := strconv.Atoi(r.FormValue("die_index"))
	room.Dice.ToggleDie(dieIdx)
	room.Broadcaster.Broadcast(broadcaster.Event{Name: "diceAreaUpdated"})
	room.Broadcaster.Broadcast(broadcaster.Event{Name: "scoreUpdated"})

	playCookie, err := r.Cookie("player_id")
	if err != nil {
		http.Error(w, "no player cookie", http.StatusForbidden)
		log.Println("no player cookie:", err)
		return
	}
	playerID := playCookie.Value

	err = views.DiceArea(roomID, playerID, room).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "could not render dice area", http.StatusInternalServerError)
		log.Println("error rendering dice area:", err)
		return
	}
}

func SelectCellHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.FormValue("room_id")
	roomsMu.Lock()
	room, ok := rooms[roomID]
	roomsMu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	playerCookie, err := r.Cookie("player_id")
	if err != nil {
		http.Error(w, "no player cookie", http.StatusForbidden)
		log.Println("no player cookie:", err)
		return
	}
	playerID := playerCookie.Value

	player := room.GetPlayerByID(playerID)
	if player == nil {
		http.Error(w, "player not in room", http.StatusForbidden)
		return
	}

	if room.Players[room.CurrentTurn].ID != playerID {
		http.Error(w, "not your turn", http.StatusForbidden)
		return
	}

	row := r.FormValue("row")
	col := r.FormValue("col")

	_, err = player.ScoreCard.FillField(row, col, room.Dice)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not fill cell: %v", err), http.StatusBadRequest)
		log.Println("error filling cell:", err)
		return
	}

	room.EndTurn() // end the turn when the user enters result in a cell
	room.Broadcaster.Broadcast(broadcaster.Event{Name: "turnEnded"})
	room.Broadcaster.Broadcast(broadcaster.Event{Name: "scoreUpdated"})

	err = views.ScoreCardField(roomID, player.ScoreCard, row, col).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "could not render score", http.StatusInternalServerError)
		log.Println("error rendering score:", err)
		return
	}
}

func OtherScorecardsHandler(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	roomsMu.Lock()
	room, ok := rooms[roomID]
	roomsMu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	playerCookie, err := r.Cookie("player_id")
	if err != nil {
		http.Error(w, "no player cookie", http.StatusForbidden)
		log.Println("no player cookie:", err)
		return
	}
	playerID := playerCookie.Value

	if err := views.OtherScorecards(roomID, playerID, room).Render(r.Context(), w); err != nil {
		http.Error(w, "could not render scorecards", http.StatusInternalServerError)
		log.Println("error rendering scorecards:", err)
		return
	}
}

func EventsHandler(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")

	roomsMu.Lock()
	room, ok := rooms[roomID]
	roomsMu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch := room.Broadcaster.Subscribe()
	defer room.Broadcaster.Unsubscribe(ch)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-ch:
			fmt.Fprintf(w, "event: %s\n", ev.Name)
			fmt.Fprintf(w, "data: _\n\n")
			flusher.Flush()
		}
	}
}

func DiceAreaHandler(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	roomsMu.Lock()
	room, ok := rooms[roomID]
	roomsMu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	playerCookie, err := r.Cookie("player_id")
	if err != nil {
		http.Error(w, "no player cookie", http.StatusForbidden)
		log.Println("no player cookie:", err)
		return
	}
	playerID := playerCookie.Value

	err = views.DiceArea(roomID, playerID, room).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "could not render dice area", http.StatusInternalServerError)
		log.Println("error rendering dice area:", err)
		return
	}
}

func PlayerCounterHandler(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	roomsMu.Lock()
	room, ok := rooms[roomID]
	roomsMu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	err := views.PlayerCounter(room).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "could not render player counter", http.StatusInternalServerError)
		log.Println("error rendering player counter:", err)
		return
	}
}
