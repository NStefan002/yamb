package main

import (
	"bytes"
	"context"
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
	"golang.org/x/net/websocket"
)

var (
	rooms   = make(map[string]*game.Room)
	roomsMu sync.Mutex
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	err := views.Index().Render(r.Context(), w)
	if err != nil {
		HxError(w, "could not render index", http.StatusInternalServerError)
		log.Println("error rendering index:", err)
		return
	}
}

func CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		HxError(w, "bad form", http.StatusBadRequest)
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
		HxError(w, "could not render room link", http.StatusInternalServerError)
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
		HxError(w, "room does not exist", 404)
		return
	}

	err := views.UsernameEntry(roomID).Render(r.Context(), w)
	if err != nil {
		HxError(w, "could not render username entry", http.StatusInternalServerError)
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
		HxError(w, "room does not exist", 404)
		return
	}

	if room.IsFull() {
		HxError(w, "room full", http.StatusForbidden)
		return
	}

	playerID := uuid.New().String()
	http.SetCookie(w, &http.Cookie{
		Name:  "player_id",
		Value: playerID,
		Path:  "/",
	})
	// TODO: do we want to use this cookie instead of passing room_id in forms?
	http.SetCookie(w, &http.Cookie{
		Name:  "room_id",
		Value: roomID,
		Path:  "/",
	})

	err := room.AddPlayer(game.NewPlayer(playerID, username))
	if err != nil {
		HxError(w, fmt.Sprintf("could not add player: %v", err), http.StatusInternalServerError)
		log.Println("error adding player to room:", err)
		return
	}

	room.Broadcaster.Broadcast(broadcaster.Event{Name: broadcaster.PlayerJoined})
	room.Broadcaster.Broadcast(broadcaster.Event{Name: broadcaster.ScoreUpdated})

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
		HxError(w, "room does not exist", 404)
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
		HxError(w, "could not render room page", http.StatusInternalServerError)
		log.Println("error rendering room page:", err)
		return
	}
}

func ResultsPageHandler(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	roomsMu.Lock()
	room, ok := rooms[roomID]
	roomsMu.Unlock()
	if !ok {
		HxError(w, "room does not exist", 404)
		return
	}

	playCookie, err := r.Cookie("player_id")
	if err != nil {
		HxError(w, "no player cookie", http.StatusForbidden)
		log.Println("no player cookie:", err)
		return
	}
	playerID := playCookie.Value

	err = views.ResultsPage(roomID, playerID, room).Render(r.Context(), w)
	if err != nil {
		HxError(w, "could not render results page", http.StatusInternalServerError)
		log.Println("error rendering results page:", err)
		return
	}
}

func RollDiceHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.FormValue("room_id")
	roomsMu.Lock()
	room, ok := rooms[roomID]
	roomsMu.Unlock()
	if !ok {
		HxError(w, "room does not exist", 404)
		return
	}

	room.RollDice()
	room.Broadcaster.Broadcast(broadcaster.Event{Name: broadcaster.DiceAreaUpdated})
	room.Broadcaster.Broadcast(broadcaster.Event{Name: broadcaster.ScoreUpdated})

	playerCookie, err := r.Cookie("player_id")
	if err != nil {
		HxError(w, "no player cookie", http.StatusForbidden)
		log.Println("no player cookie:", err)
		return
	}
	playerID := playerCookie.Value

	err = views.DiceArea(roomID, playerID, room).Render(r.Context(), w)
	if err != nil {
		HxError(w, "could not render dice area", http.StatusInternalServerError)
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
		HxError(w, "room does not exist", 404)
		return
	}

	dieIdx, _ := strconv.Atoi(r.FormValue("die_index"))
	room.Dice.ToggleDie(dieIdx)
	room.Broadcaster.Broadcast(broadcaster.Event{Name: broadcaster.DiceAreaUpdated})
	room.Broadcaster.Broadcast(broadcaster.Event{Name: broadcaster.ScoreUpdated})

	playCookie, err := r.Cookie("player_id")
	if err != nil {
		HxError(w, "no player cookie", http.StatusForbidden)
		log.Println("no player cookie:", err)
		return
	}
	playerID := playCookie.Value

	err = views.DiceArea(roomID, playerID, room).Render(r.Context(), w)
	if err != nil {
		HxError(w, "could not render dice area", http.StatusInternalServerError)
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
		HxError(w, "room does not exist", 404)
		return
	}

	playerCookie, err := r.Cookie("player_id")
	if err != nil {
		HxError(w, "no player cookie", http.StatusForbidden)
		log.Println("no player cookie:", err)
		return
	}
	playerID := playerCookie.Value

	player := room.GetPlayerByID(playerID)
	if player == nil {
		HxError(w, "player not in room", http.StatusForbidden)
		return
	}

	if room.Players[room.CurrentTurn].ID != playerID {
		HxError(w, "not your turn", http.StatusForbidden)
		return
	}

	row := r.FormValue("row")
	col := r.FormValue("col")

	err = player.ScoreCard.SelectCell(row, col)
	if err != nil {
		HxError(w, fmt.Sprintf("could not select cell: %v", err), http.StatusBadRequest)
		log.Println("error selecting cell:", err)
		return
	}

	room.Broadcaster.Broadcast(broadcaster.Event{Name: broadcaster.CellSelected})

	err = views.MainScoreCard(roomID, playerID, room).Render(r.Context(), w)
	if err != nil {
		HxError(w, "could not render score", http.StatusInternalServerError)
		log.Println("error rendering score:", err)
		return
	}
}

func WriteScoreHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.FormValue("room_id")
	roomsMu.Lock()
	room, ok := rooms[roomID]
	roomsMu.Unlock()
	if !ok {
		HxError(w, "room does not exist", 404)
		return
	}

	playerCookie, err := r.Cookie("player_id")
	if err != nil {
		HxError(w, "no player cookie", http.StatusForbidden)
		log.Println("no player cookie:", err)
		return
	}
	playerID := playerCookie.Value

	player := room.GetPlayerByID(playerID)
	if player == nil {
		HxError(w, "player not in room", http.StatusForbidden)
		return
	}

	if room.Players[room.CurrentTurn].ID != playerID {
		HxError(w, "not your turn", http.StatusForbidden)
		return
	}

	announce := r.FormValue("announce") == "true"

	if announce {
		player.ScoreCard.Announce()
		room.Broadcaster.Broadcast(broadcaster.Event{Name: broadcaster.ScoreAnnounced})
		err = views.MainScoreCard(roomID, playerID, room).Render(r.Context(), w)
		if err != nil {
			HxError(w, "could not render score", http.StatusInternalServerError)
			log.Println("error rendering score:", err)
			return
		}
	} else {
		row, col := player.ScoreCard.GetSelectedCell()

		_, err = player.ScoreCard.FillCell(row, col, room.Dice)
		if err != nil {
			HxError(w, fmt.Sprintf("could not fill cell: %v", err), http.StatusBadRequest)
			log.Println("error filling cell:", err)
			return
		}

		player.ScoreCard.CalculateSums()

		room.EndTurn() // end the turn when the user enters result in a cell
        player.ScoreCard.UnselectCell()
		room.Broadcaster.Broadcast(broadcaster.Event{Name: broadcaster.TurnEnded})
		room.Broadcaster.Broadcast(broadcaster.Event{Name: broadcaster.ScoreUpdated})

		err = views.MainScoreCard(roomID, playerID, room).Render(r.Context(), w)
		if err != nil {
			HxError(w, "could not render score", http.StatusInternalServerError)
			log.Println("error rendering score:", err)
			return
		}

		if room.GameEnded() {
			room.SortPlayersByScore()
			room.Broadcaster.Broadcast(broadcaster.Event{Name: broadcaster.GameEnded})
			return
		}
	}
}

func OtherScorecardsHandler(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	roomsMu.Lock()
	room, ok := rooms[roomID]
	roomsMu.Unlock()
	if !ok {
		HxError(w, "room does not exist", 404)
		return
	}

	playerCookie, err := r.Cookie("player_id")
	if err != nil {
		HxError(w, "no player cookie", http.StatusForbidden)
		log.Println("no player cookie:", err)
		return
	}
	playerID := playerCookie.Value

	if err := views.OtherScorecards(roomID, playerID, room).Render(r.Context(), w); err != nil {
		HxError(w, "could not render scorecards", http.StatusInternalServerError)
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
		HxError(w, "room does not exist", 404)
		return
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		HxError(w, "Streaming unsupported", http.StatusInternalServerError)
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
		HxError(w, "room does not exist", 404)
		return
	}

	playerCookie, err := r.Cookie("player_id")
	if err != nil {
		HxError(w, "no player cookie", http.StatusForbidden)
		log.Println("no player cookie:", err)
		return
	}
	playerID := playerCookie.Value

	err = views.DiceArea(roomID, playerID, room).Render(r.Context(), w)
	if err != nil {
		HxError(w, "could not render dice area", http.StatusInternalServerError)
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
		HxError(w, "room does not exist", 404)
		return
	}

	err := views.PlayerCounter(room).Render(r.Context(), w)
	if err != nil {
		HxError(w, "could not render player counter", http.StatusInternalServerError)
		log.Println("error rendering player counter:", err)
		return
	}
}

func CellSelectedHandler(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	roomsMu.Lock()
	room, ok := rooms[roomID]
	roomsMu.Unlock()
	if !ok {
		HxError(w, "room does not exist", 404)
		return
	}

	playerCookie, err := r.Cookie("player_id")
	if err != nil {
		HxError(w, "no player cookie", http.StatusForbidden)
		log.Println("no player cookie:", err)
		return
	}
	playerID := playerCookie.Value

	err = views.WriteScoreButton(roomID, playerID, room).Render(r.Context(), w)
	if err != nil {
		HxError(w, "could not render write score button", http.StatusInternalServerError)
		log.Println("error rendering write score button:", err)
		return
	}
}

func HxError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"showError": %q}`, msg))
	w.WriteHeader(status)
}

func ChatWebsocketHandler(ws *websocket.Conn) {
	roomID := chi.URLParam(ws.Request(), "roomID")
	roomsMu.Lock()
	room, ok := rooms[roomID]
	roomsMu.Unlock()
	if !ok {
		ws.Close()
		return
	}

	room.Mu.Lock()
	room.ChatConns[ws] = true
	room.Mu.Unlock()

	defer func() {
		room.RemoveConn(ws)
		ws.Close()
	}()

	for {
		var msg struct {
			Msg      string `json:"msg"`
			PlayerID string `json:"player_id"`
		}
		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			break
		}
		player := room.GetPlayerByID(msg.PlayerID)
		username := "Unknown"
		if player != nil {
			username = player.Username
		}
		// Render chat message HTML
		var buf bytes.Buffer
		err := views.ChatMessage(username, msg.Msg).Render(context.Background(), &buf)
		if err != nil {
			log.Println("error rendering chat message:", err)
			continue
		}
		room.Broadcast(buf.String())
	}
}
