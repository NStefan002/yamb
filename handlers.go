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
	lang := getLang(r)

	err := views.Index(lang).Render(r.Context(), w)
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

	lang := getLang(r)

	err := views.RoomLink(roomID, lang).Render(r.Context(), w)
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

	lang := getLang(r)

	err := views.UsernameEntry(roomID, lang).Render(r.Context(), w)
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

	lang := getLang(r)

	err = views.RoomPage(roomID, playerID, lang, room).Render(r.Context(), w)
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

	lang := getLang(r)

	err = views.ResultsPage(roomID, playerID, lang, room).Render(r.Context(), w)
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

	lang := getLang(r)

	err = views.DiceArea(roomID, playerID, lang, room).Render(r.Context(), w)
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

	lang := getLang(r)

	err = views.DiceArea(roomID, playerID, lang, room).Render(r.Context(), w)
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

	lang := getLang(r)

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

	err = views.MainScoreCard(roomID, playerID, lang, room).Render(r.Context(), w)
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

	lang := getLang(r)

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
		err = views.MainScoreCard(roomID, playerID, lang, room).Render(r.Context(), w)
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

		err = views.MainScoreCard(roomID, playerID, lang, room).Render(r.Context(), w)
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

	lang := getLang(r)

	if err := views.OtherScorecards(roomID, playerID, lang, room).Render(r.Context(), w); err != nil {
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

	lang := getLang(r)

	err = views.DiceArea(roomID, playerID, lang, room).Render(r.Context(), w)
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

	lang := getLang(r)

	err := views.PlayerCounter(lang, room).Render(r.Context(), w)
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

	lang := getLang(r)

	err = views.WriteScoreButton(roomID, playerID, lang, room).Render(r.Context(), w)
	if err != nil {
		HxError(w, "could not render write score button", http.StatusInternalServerError)
		log.Println("error rendering write score button:", err)
		return
	}
}

func SetLangHandler(w http.ResponseWriter, r *http.Request) {
	lang := r.FormValue("lang")
	if lang == "" {
		http.Error(w, "missing lang", http.StatusBadRequest)
		return
	}

	// create lang cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "lang",
		Value:    lang,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// return no content
	w.WriteHeader(http.StatusNoContent)
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
		chatMsg := game.NewChatMessage(player.ID, msg.Msg)
		// add message to room chat history
		room.Mu.Lock()
		room.ChatHistory = append(room.ChatHistory, chatMsg)
		room.Mu.Unlock()
		// Render chat message HTML
		var buf bytes.Buffer
		err := views.ChatMessageWrapper(player, chatMsg).Render(context.Background(), &buf)
		if err != nil {
			log.Println("error rendering chat message:", err)
			break
		}
		room.Broadcast(buf.String())
	}
}

func getLang(r *http.Request) string {
	langCookie, err := r.Cookie("lang")
	if err != nil {
		log.Println("no lang cookie:", err)
		return "en"
	}
	return langCookie.Value
}
