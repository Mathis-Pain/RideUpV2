package handlers

import (
	"RideUP/models"
	"RideUP/sessions"
	"RideUP/utils"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

var PreferenceHtml = template.Must(template.ParseFiles(
	"templates/preference.html",
	"templates/inithtml/inithead.html",
	"templates/inithtml/initnav.html",
	"templates/inithtml/initfooter.html",
))

func PreferenceHandler(w http.ResponseWriter, r *http.Request) {

	// Récupère l'utilisateur connecté
	session, err := sessions.GetSessionFromRequest(r)
	if err != nil {
		log.Printf("Erreur pas d'utilisateur connecté")
		http.Redirect(w, r, "/Connect", http.StatusSeeOther)
		return
	}

	db, err := sql.Open("sqlite3", "./data/RideUp.db")
	if err != nil {
		utils.InternalServError(w)
		return
	}
	defer db.Close()

	// Coordonnées par défaut ou de l'utilisateur
	var data models.MapData
	err = db.QueryRow(`SELECT latitude, longitude, address FROM users WHERE id = ?`, session.UserID).
		Scan(&data.Latitude, &data.Longitude, &data.Address)
	if err != nil {
		log.Println("Erreur récupération coordonnées:", err)
		data.Latitude = 48.8566
		data.Longitude = 2.3522
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			log.Printf("ERREUR : ParseForm: %v", err)
			http.Error(w, "Erreur lors de la lecture du formulaire", http.StatusBadRequest)
			return
		}

		addressuser := r.FormValue("addressuser") // Adresse (peut être vide si coords directes)
		latStr := r.FormValue("latitude")         // Champs hidden carte
		lonStr := r.FormValue("longitude")        // Champs hidden carte
		radius := r.FormValue("radius")

		// Vérifier qu'on a soit une adresse, soit des coordonnées
		if addressuser == "" && (latStr == "" || lonStr == "") {
			http.Error(w, "Veuillez saisir une adresse ou sélectionner un point sur la carte", http.StatusBadRequest)
			return
		}
		// Récupération des coordonnées et de l'adresse finale
		var lat, lon float64
		var finalAddress string

		if latStr != "" && lonStr != "" {
			// Coordonnées via carte (prioritaire)
			lat, err = strconv.ParseFloat(latStr, 64)
			if err != nil {
				log.Printf("Erreur parsing latitude: %v", err)
				http.Error(w, "Coordonnées invalides", http.StatusBadRequest)
				return
			}
			lon, err = strconv.ParseFloat(lonStr, 64)
			if err != nil {
				log.Printf("Erreur parsing longitude: %v", err)
				http.Error(w, "Coordonnées invalides", http.StatusBadRequest)
				return
			}
			// Si une adresse est fournie (via géocodage inversé JS), on l'utilise
			if addressuser != "" {
				finalAddress = addressuser
			} else {
				finalAddress = fmt.Sprintf("Lat: %.6f, Lon: %.6f", lat, lon)
			}

		} else if addressuser != "" {
			// Géocodage adresse uniquement si pas de coordonnées
			lat, lon, err = utils.GeocodeAddress(addressuser)
			if err != nil {
				log.Printf("Erreur géocodage : %v", err)
				http.Error(w, "Adresse introuvable. Veuillez vérifier l'orthographe ou utiliser la carte.", http.StatusBadRequest)
				return
			}
			finalAddress = addressuser
		}
		// Insertion dans la base de données avec l'adresse
		_, err = db.Exec(`
    UPDATE users 
    SET latitude = ?, longitude = ?, address = ?, preference = ?
    WHERE id = ?`,
			lat, lon, finalAddress, radius, session.UserID)
		if err != nil {
			log.Printf("ERREUR : insertion user: %v", err)
			http.Error(w, "Impossible de modifier l'adresse", http.StatusInternalServerError)
			return
		}
		log.Printf("✅ Adresse modifier avec succes")
		http.Redirect(w, r, "/RideUp", http.StatusSeeOther)
		return
	}
	if err := PreferenceHtml.Execute(w, data); err != nil {
		log.Printf("Erreur lors de l'exécution du template rideup.html: %v", err)
		utils.InternalServError(w)
	}
}
