package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"RideUP/handlers"
	"RideUP/hub"
	"RideUP/middleware"
)

func main() {
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		log.Fatal("SUPABASE_URL requis")
	}

	if err := middleware.Init(supabaseURL); err != nil {
		log.Fatalf("JWKS init: %v", err)
	}

	h := hub.New()
	go h.Run()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	// WebSocket chat par event — auth via ?token=<jwt>
	mux.Handle("/ws/chat/{eventId}", middleware.Auth(handlers.ChatWS(h)))

	// Requis pour la persistance des messages
	if os.Getenv("SUPABASE_ANON_KEY") == "" {
		log.Println("WARN: SUPABASE_ANON_KEY non défini, persistance désactivée")
	}

	log.Println("API démarrée sur :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
