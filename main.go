package main

import (
	builddb "RideUP/buildDB"
	"RideUP/routes"
	"RideUP/sessions"
	"RideUP/utils/authextern"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// initialisation de la bdd
	db, err := builddb.InitDB()
	if err != nil {
		fmt.Println("Erreur creation bdd :", err)
		return
	}
	defer db.Close()
	fmt.Println("Projet lancé, DB prête a l'emploi")
	// Identification google
	authextern.InitGoogleOAuth()
	// Nettoyage des sessions expirées toutes les 5 minutes
	go func() {
		for {
			time.Sleep(30 * time.Minute)
			sessions.CleanupExpiredSessions()
		}
	}()

	//initialisation des routes
	mux := routes.InitRoutes()

	// Demarrage du serveur
	fmt.Println("serveur démarré sur http://localhost:5090...")
	if err := http.ListenAndServe(":5090", mux); err != nil {
		log.Fatal("Erreur serveur:", err)
	}

}
