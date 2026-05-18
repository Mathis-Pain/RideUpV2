package handlers

import (
	"RideUP/sessions"
	"RideUP/utils"
	"log"
	"net/http"
)

func DisconnectHandler(w http.ResponseWriter, r *http.Request) {

	// Récupère la session depuis la requête
	session, err := sessions.GetSessionFromRequest(r)
	if err != nil {
		log.Println("ERREUR : <logouthandler.go> Erreur lors de la récupération de la session :", err)
		utils.InternalServError(w)
		return
	}

	// Supprime le cookie de session côté navigateur
	sessions.DeleteCookie(w, "session_id", false) // false si local, true si HTTPS
	// Supprime la session côté serveur/DB
	sessions.DeleteSession(session.ID)

	// Redirige vers la page d’accueil
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
