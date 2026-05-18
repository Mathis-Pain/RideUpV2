package handlers

import (
	"RideUP/sessions"
	"RideUP/utils"
	"database/sql"
	"html/template"
	"log"
	"net/http"
)

// Template de connexion
var ConnectionHtml = template.Must(template.ParseFiles(
	"templates/connection.html",
	"templates/inithtml/inithead.html",
	"templates/inithtml/initfooter.html",
))

// Struct pour passer les données au template
type LoginFormData struct {
	Email string
	Error string
}

// ConnectHandler gère la page de connexion
func ConnectHandler(w http.ResponseWriter, r *http.Request) {
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/"
	}

	// Vérifier si la session est active
	sessionCookie, err := r.Cookie("session_id")
	if err == nil && sessions.IsValidSession(sessionCookie.Value) {
		http.Redirect(w, r, "/RideUp", http.StatusSeeOther)
		return
	}

	// ----------------- GET -----------------
	if r.Method == http.MethodGet {
		if err := ConnectionHtml.Execute(w, nil); err != nil {
			log.Printf("Erreur affichage page connexion: %v", err)
			utils.InternalServError(w)
			return
		}
		return
	}

	// ----------------- POST -----------------
	if r.Method == http.MethodPost {
		// Récupérer les valeurs du formulaire
		if err := r.ParseForm(); err != nil {
			log.Printf("Erreur ParseForm: %v", err)
			http.Error(w, "Erreur lors de la lecture du formulaire", http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		// Ouvrir la base de données
		db, err := sql.Open("sqlite3", "./data/RideUp.db")
		if err != nil {
			utils.InternalServError(w)
			log.Printf("Erreur ouverture base de données: %v", err)
			return
		}
		defer db.Close()

		// Authentification
		user, loginErr := utils.Authentification(db, email, password)
		if loginErr != nil {
			// Afficher l'erreur dans le formulaire
			data := LoginFormData{
				Email: email,
				Error: loginErr.Error(),
			}
			w.WriteHeader(http.StatusBadRequest)
			if err := ConnectionHtml.Execute(w, data); err != nil {
				log.Printf("Erreur affichage template avec erreur login: %v", err)
			}
			return
		}

		// Si tout est OK, invalider les anciennes sessions et créer la nouvelle
		if err := sessions.InvalidateUserSessions(user.ID); err != nil {
			utils.InternalServError(w)
			return
		}
		if err := InitSession(w, user.ID, "user", user.Email); err != nil {
			utils.InternalServError(w)
			return
		}

		// Redirection vers le tableau de bord
		http.Redirect(w, r, "/RideUp", http.StatusSeeOther)
		return
	}

	// Méthode HTTP non autorisée
	http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
}

// InitSession crée et sauvegarde une session puis pose le cookie
func InitSession(w http.ResponseWriter, id int, fieldName string, fieldData any) error {
	session, err := sessions.CreateSession(id)
	if err != nil {
		return err
	}

	session.Data[fieldName] = fieldData
	if err := sessions.SaveSessionToDB(session); err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   false, // false en local, true si HTTPS
		Path:     "/",
	})
	return nil
}
