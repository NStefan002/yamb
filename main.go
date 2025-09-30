package main

import (
	"fmt"
	"log"
	"net/http"
	"yamb/views"

	"github.com/a-h/templ"
)

func main() {
	mux := http.NewServeMux()

	// Routes
	mux.Handle("/", templ.Handler(views.Index()))
	mux.HandleFunc("/create-room", handleCreateRoom)
	mux.HandleFunc("/room/", handleRoomPage)

	// Actions (HTMX endpoints)
	mux.HandleFunc("/roll-dice", handleRollDice)
	mux.HandleFunc("/toggle-dice", handleToggleDice)
	mux.HandleFunc("/select-cell", handleSelectCell)
	// mux.HandleFunc("/send-message", handleSendMessage)

	fmt.Println("Listening on http://localhost:1312")
	log.Fatal(http.ListenAndServe(":1312", mux))
}
