package handlers

import (
	"RideUP/models"
	"RideUP/sessions"
	"RideUP/utils"
	"RideUP/utils/getdata"
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

var UpdatePasswordHtml = template.Must(template.ParseFiles(
	"templates/updatepassword.html",
	"templates/inithtml/inithead.html",
	"templates/inithtml/initnav.html",
	"templates/inithtml/initfooter.html",
))

func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifie la session dès le début pour GET et POST
	session, err := sessions.GetSessionFromRequest(r)
	if err != nil {
		http.Redirect(w, r, "/Connect", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Affichage initial du formulaire
		if err := UpdatePasswordHtml.Execute(w, nil); err != nil {
			log.Printf("Erreur template GET UpdatePasswordHtml: %v", err)
			utils.NotFoundHandler(w)
		}
		return

	case http.MethodPost:
		// Connexion à la DB
		db, err := sql.Open("sqlite3", "./data/RideUp.db")
		if err != nil {
			log.Printf("Erreur ouverture DB: %v", err)
			utils.InternalServError(w)
			return
		}
		defer db.Close()

		// Parse le formulaire
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Erreur formulaire", http.StatusBadRequest)
			return
		}

		oldPassword := r.FormValue("OldPassword")
		newPassword := r.FormValue("NewPassword")
		confirmPassword := r.FormValue("confirmNewPassword") // ← Attention à la casse !

		// Struct pour les erreurs
		formData := models.UpdatePassword{}

		// 1. Récupère le hash actuel
		passwordHash, err := getdata.GetPasswordHash(db, session.UserID)
		if err != nil {
			log.Printf("Erreur récupération du mot de passe: %v", err)
			utils.InternalServError(w)
			return
		}

		// 2. Vérifie l'ancien mot de passe
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(oldPassword)); err != nil {
			formData.OldPasswordError = "Ancien mot de passe incorrect"
		}

		// 3. Vérifie que les nouveaux mots de passe correspondent
		if newPassword != confirmPassword {
			formData.ConfirmNewPasswordError = "Les deux nouveaux mots de passe ne sont pas identiques."
		}

		// 4. Valide le nouveau mot de passe
		if newPassErr := utils.ValidPassword(newPassword, confirmPassword); newPassErr != "" {
			formData.NewPasswordError = newPassErr
		}

		// 5. Si des erreurs existent, renvoie le formulaire
		if formData.OldPasswordError != "" || formData.NewPasswordError != "" || formData.ConfirmNewPasswordError != "" {
			if err := UpdatePasswordHtml.Execute(w, formData); err != nil {
				log.Printf("Erreur affichage template avec erreurs: %v", err)
			}
			return
		}

		// 6. Tout est OK → hash et update
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Erreur génération hash: %v", err)
			utils.InternalServError(w)
			return
		}

		_, err = db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, hashedPassword, session.UserID)
		if err != nil {
			log.Printf("Erreur de mise à jour du mot de passe: %v", err)
			utils.InternalServError(w)
			return
		}

		// 7. Succès → redirection
		http.Redirect(w, r, "/Profil", http.StatusSeeOther)

	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}
