package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"yamb/game"
	"yamb/views"
)

var rooms = make(map[string]*game.Room)

func handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	roomID := fmt.Sprintf("%06d", rand.Intn(1000000))

	rooms[roomID] = game.NewRoom()

	views.RoomLink(fmt.Sprintf("http://localhost:1312/room/%s", roomID)).Render(r.Context(), w)
}

func handleRoomPage(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Path[len("/room/"):]
	room, ok := rooms[roomID]
	if !ok {
		http.NotFound(w, r)
		return
	}

	views.RoomPage(roomID, room).Render(r.Context(), w)
}

func handleRollDice(w http.ResponseWriter, r *http.Request) {
	roomID := r.FormValue("room_id")
	room, ok := rooms[roomID]
	if !ok {
		http.NotFound(w, r)
		return
	}

	room.RollDice()
	views.DiceArea(roomID, room.Players[0].Dice).Render(r.Context(), w)
}

func handleToggleDice(w http.ResponseWriter, r *http.Request) {
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

func handleSelectCell(w http.ResponseWriter, r *http.Request) {
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
