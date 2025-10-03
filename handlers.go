package main

import (
	"fmt"
	"log"
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
		log.Println("error parsing form:", err)
		return
	}

	roomID := fmt.Sprintf("%06d", rand.Intn(1000000))
	mode := r.FormValue("mode")
	dice := r.FormValue("dice")

	rooms[roomID] = game.NewRoom(mode, dice)

	err := views.RoomLink(roomID).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "could not render room link", http.StatusInternalServerError)
		log.Println("error rendering room link:", err)
		return
	}
}

func RoomLinkHandler(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	if _, ok := rooms[roomID]; !ok {
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

	err := room.AddPlayer(game.NewPlayer(playerID, username))
	if err != nil {
		http.Error(w, fmt.Sprintf("could not add player: %v", err), http.StatusInternalServerError)
		log.Println("error adding player to room:", err)
		return
	}

	if len(room.Players) == room.NumOfPlayers {
		room.GameStarted = true
	}

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
	room, ok := rooms[roomID]
	if !ok {
		http.NotFound(w, r)
		return
	}

	room.RollDice()
	fmt.Println("TURN: ", room.CurrentTurn)
	err := views.DiceArea(roomID, room.Dice).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "could not render dice area", http.StatusInternalServerError)
		log.Println("error rendering dice area:", err)
		return
	}
}

func ToggleDiceHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.FormValue("room_id")
	room, ok := rooms[roomID]
	if !ok {
		http.NotFound(w, r)
		return
	}

	dieIdx, _ := strconv.Atoi(r.FormValue("die_index"))
	room.Dice.ToggleDie(dieIdx)
	err := views.DiceArea(roomID, room.Dice).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "could not render dice area", http.StatusInternalServerError)
		log.Println("error rendering dice area:", err)
		return
	}
}

func SelectCellHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.FormValue("room_id")
	room, ok := rooms[roomID]
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

	err = views.ScoreCardField(roomID, player.ScoreCard, row, col).Render(r.Context(), w)
	if err != nil {
		http.Error(w, "could not render score", http.StatusInternalServerError)
		log.Println("error rendering score:", err)
		return
	}
}
