package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"yamb/game"
	"yamb/views"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var rooms = make(map[string]*game.Room)

func CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	roomID := fmt.Sprintf("%06d", rand.Intn(1000000))

	rooms[roomID] = game.NewRoom()

	views.RoomLink(roomID).Render(r.Context(), w)
}

func RoomLinkHandler(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	if _, ok := rooms[roomID]; !ok {
		http.NotFound(w, r)
		return
	}

	views.UsernameEntry(roomID).Render(r.Context(), w)
}

func JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.FormValue("room_id")
	username := r.FormValue("username")

	room, ok := rooms[roomID]
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

	room.AddPlayer(game.NewPlayer(playerID, username))

	http.Redirect(w, r, fmt.Sprintf("/room/%s", roomID), http.StatusSeeOther)
}

func RoomPageHandler(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	room, ok := rooms[roomID]
	if !ok {
		http.NotFound(w, r)
		return
	}

	playerCookie, err := r.Cookie("player_id")
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/room/%s", roomID), http.StatusSeeOther)
		return
	}

	playerID := playerCookie.Value
	views.RoomPage(roomID, playerID, room).Render(r.Context(), w)
}

func RollDiceHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.FormValue("room_id")
	room, ok := rooms[roomID]
	if !ok {
		http.NotFound(w, r)
		return
	}

	room.RollDice()
	views.DiceArea(roomID, room.Players[0].Dice).Render(r.Context(), w)
}

func ToggleDiceHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.FormValue("room_id")
	room, ok := rooms[roomID]
	if !ok {
		http.NotFound(w, r)
		return
	}

	dieIdx, _ := strconv.Atoi(r.FormValue("die_index"))
	room.Players[0].Dice.ToggleDie(dieIdx)
	views.DiceArea(roomID, room.Players[0].Dice).Render(r.Context(), w)
}

func SelectCellHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room_id")
	room, ok := rooms[roomID]
	if !ok {
		http.NotFound(w, r)
		return
	}

	row := r.URL.Query().Get("row")
	col := r.URL.Query().Get("col")

	// for testing purposes, just set a random number
	score := rand.Intn(30) + 1

	room.Players[0].ScoreCard.Scores[row][col] = &score

	fmt.Fprintf(w, `<td class="border border-gray-400 w-12 h-8 bg-green-100 text-center">%d</td>`, score)

	fmt.Println("Cell selected", row, col, "=", score)
}
