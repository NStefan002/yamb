package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/net/websocket"

	"yamb/i18n"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// load js library files
	jsFiles := http.StripPrefix("/js/", http.FileServer(http.Dir("assets/js")))
	r.Handle("/js/*", jsFiles)

	// load css files
	cssFiles := http.StripPrefix("/css/", http.FileServer(http.Dir("assets/css")))
	r.Handle("/css/*", cssFiles)

	// load locales
	localeFiles := http.StripPrefix("/locales/", http.FileServer(http.Dir("assets/locales")))
	r.Handle("/locales/*", localeFiles)

	err := i18n.LoadLocales("assets/locales")
	if err != nil {
		log.Fatal(err)
	}

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

	// results page
	r.Get("/room/{roomID}/results", ResultsPageHandler)

	// change language
	r.Post("/set-lang", SetLangHandler)

	// HTMX endpoints for partial updates (events)
	r.Get("/room/{roomID}/events", EventsHandler)
	r.Get("/room/{roomID}/other-scorecards", OtherScorecardsHandler)
	r.Get("/room/{roomID}/dice-area", DiceAreaHandler)
	r.Get("/room/{roomID}/player-counter", PlayerCounterHandler)
	r.Get("/room/{roomID}/cell-selected", CellSelectedHandler)

	// Actions (HTMX endpoints)
	r.Post("/roll-dice", RollDiceHandler)
	r.Post("/toggle-dice", ToggleDiceHandler)
	r.Post("/select-cell", SelectCellHandler)
	r.Post("/write-score", WriteScoreHandler)

	// Chat endpoints
	r.Handle("/room/{roomID}/chat/", websocket.Handler(ChatWebsocketHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	port = ":" + port
	log.Fatal(http.ListenAndServe(port, r))
}
