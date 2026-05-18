package middleware

import (
	"RideUP/sessions"
	"RideUP/utils/getdata"
	"database/sql"
	"log"
	"net/http"
)

// AdminOnly vérifie si l'utilisateur connecté est admin
func AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Récupération de la session
		session, err := sessions.GetSessionFromRequest(r)
		if err != nil {
			log.Printf("Erreur : pas d'utilisateur connecté")
			http.Redirect(w, r, "/Connect", http.StatusSeeOther)
			return
		}

		// Connexion à la DB
		db, err := sql.Open("sqlite3", "./data/RideUp.db")
		if err != nil {
			log.Printf("Erreur DB dans middleware AdminOnly: %v", err)
			http.Redirect(w, r, "/RideUp", http.StatusSeeOther)
			return
		}
		defer db.Close()

		// Vérification du statut admin
		isAdmin, err := getdata.IsUserAdmin(db, session.UserID)
		if err != nil || !isAdmin {
			log.Printf("Accès refusé : utilisateur %d n'est pas admin", session.UserID)
			http.Redirect(w, r, "/RideUp", http.StatusSeeOther)
			return
		}

		// L'utilisateur est admin, on continue
		next(w, r)
	}
}
