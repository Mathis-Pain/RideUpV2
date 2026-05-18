package handlers

import (
	"RideUP/models"
	"RideUP/utils"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var RegistrationHtml = template.Must(template.ParseFiles("templates/registration.html", "templates/inithtml/inithead.html", "templates/inithtml/initfooter.html"))

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Gestion de la méthode GET
	if r.Method == "GET" {
		if err := RegistrationHtml.Execute(w, nil); err != nil {
			log.Printf("ERREUR : Affichage page inscription: %v", err)
			utils.InternalServError(w)
			return
		}
		return
	}

	// Gestion de la méthode POST
	if r.Method == "POST" {
		// Ouverture de la DB
		db, err := sql.Open("sqlite3", "./data/RideUp.db")
		if err != nil {
			log.Printf("ERREUR : Ouverture base de données: %v", err)
			utils.InternalServError(w)
			return
		}
		defer db.Close()

		// Récupération des valeurs du formulaire
		if err := r.ParseForm(); err != nil {
			log.Printf("ERREUR : ParseForm: %v", err)
			http.Error(w, "Erreur lors de la lecture du formulaire", http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		passwordConfirm := r.FormValue("confirmpassword")

		// Validation des données
		formData := models.RegisterDataError{
			NameError:  utils.ValidName(username),
			EmailError: utils.ValidEmail(email),
			PassError:  utils.ValidPassword(password, passwordConfirm),
		}

		data := struct {
			LoginErr     string
			RegisterData models.RegisterDataError
		}{
			LoginErr:     "",
			RegisterData: formData,
		}

		// Si une erreur de validation existe
		if formData.NameError != "" || formData.EmailError != "" || formData.PassError != "" {
			log.Printf("VALIDATION ÉCHOUÉE - Renvoi du formulaire")
			w.WriteHeader(http.StatusBadRequest)
			if err := RegistrationHtml.Execute(w, data); err != nil {
				log.Printf("ERREUR : Execute template avec erreurs: %v", err)
			}
			return
		}

		// Hash du mot de passe
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("ERREUR : Hash mot de passe: %v", err)
			utils.InternalServError(w)
			return
		}

		// Vérifie si c'est le premier utilisateur
		var count int
		role := 3 // Role par défaut (utilisateur normal)

		err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("ERREUR : Comptage utilisateurs: %v", err)
			utils.InternalServError(w)
			return
		}

		// Premier utilisateur = Admin
		if count == 0 {
			role = 1
			log.Printf("INFO : Premier utilisateur - Attribution du rôle Admin")
		}

		// Insertion dans la base de données
		log.Printf("INFO : Tentative d'insertion - username='%s', email='%s', role=%d", username, email, role)

		_, err = db.Exec("INSERT INTO users(username, email, password_hash, role_id) VALUES (?, ?, ?, ?)",
			username, email, hashedPassword, role)

		if err != nil {
			log.Printf("ERREUR INSERT: %v", err)

			// Gestion des contraintes UNIQUE
			if strings.Contains(err.Error(), "UNIQUE constraint failed: users.username") {
				formData.NameError = "Ce nom d'utilisateur est déjà pris"
				log.Printf("INFO : Username déjà existant: %s", username)
			} else if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
				formData.EmailError = "Cette adresse email est déjà utilisée"
				log.Printf("INFO : Email déjà existant: %s", email)
			} else {
				// Erreur de base de données non gérée
				log.Printf("ERREUR DB non gérée: %v", err)
				utils.InternalServError(w)
				return
			}

			// Renvoyer le formulaire avec l'erreur
			w.WriteHeader(http.StatusBadRequest)
			data.RegisterData = formData
			if err := RegistrationHtml.Execute(w, data); err != nil {
				log.Printf("ERREUR : Execute template après erreur DB: %v", err)
			}
			return
		}

		// Succès : redirection vers la page d'accueil
		log.Printf("SUCCÈS : Nouvel utilisateur inscrit: %s (role: %d)", username, role)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Méthode non autorisée
	http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
}
