package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// load htmx from assets/js
	fs := http.StripPrefix("/js/", http.FileServer(http.Dir("assets/js")))
	r.Handle("/js/*", fs)

	// landing page
	r.Get("/", IndexHandler)

	// create a room (POST from index form)
	r.Post("/create-room", CreateRoomHandler)

	// show username entry when someone visits the room link
	r.Get("/{roomID}", RoomLinkHandler)

	// join a room (username form POST)
	r.Post("/join-room", JoinRoomHandler)

	// actual game page
	r.Get("/room/{roomID}", RoomPageHandler)

	// HTMX endpoints for partial updates (events)
	r.Get("/room/{roomID}/events", EventsHandler)
	r.Get("/room/{roomID}/other-scorecards", OtherScorecardsHandler)
	r.Get("/room/{roomID}/dice-area", DiceAreaHandler)
	r.Get("/room/{roomID}/player-counter", PlayerCounterHandler)

	// Actions (HTMX endpoints)
	r.Post("/roll-dice", RollDiceHandler)
	r.Post("/toggle-dice", ToggleDiceHandler)
	r.Post("/select-cell", SelectCellHandler)
	// r.Post("/send-message", handleSendMessage)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf("http://localhost:%s", port)
	fmt.Println("Starting server on:", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
