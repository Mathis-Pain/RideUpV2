package handlers

import (
	"RideUP/models"
	"RideUP/sessions"
	"RideUP/utils"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

var NewEventHtml = template.Must(template.ParseFiles(
	"templates/newevent.html",
	"templates/inithtml/inithead.html",
	"templates/inithtml/initnav.html",
	"templates/inithtml/initfooter.html",
))

func NewEventHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier user connecté
	session, err := sessions.GetSessionFromRequest(r)
	if err != nil {
		log.Printf("Erreur : aucun utilisateur connecté")
		http.Redirect(w, r, "/Connect", http.StatusSeeOther)
		return
	}

	// DB
	db, err := sql.Open("sqlite3", "./data/RideUp.db")
	if err != nil {
		utils.InternalServError(w)
		return
	}
	defer db.Close()

	// Coordonnées par défaut utilisateur
	var data models.MapData
	err = db.QueryRow(`SELECT latitude, longitude, address FROM users WHERE id = ?`, session.UserID).
		Scan(&data.Latitude, &data.Longitude, &data.Address)
	if err != nil {
		log.Println("Erreur récupération coordonnées:", err)
		data.Latitude = 48.8566
		data.Longitude = 2.3522
		data.Address = "22 Place Saint-Marc, 76000 Rouen, France"
	}

	// ---------------------------
	// FORMULAIRE POST
	// ---------------------------
	if r.Method == http.MethodPost {

		if err := r.ParseForm(); err != nil {
			log.Printf("ERREUR ParseForm: %v", err)
			http.Error(w, "Erreur formulaire", http.StatusBadRequest)
			return
		}

		title := r.FormValue("title")
		dateStr := r.FormValue("date")
		timeStr := r.FormValue("time")
		address := r.FormValue("address")
		latStr := r.FormValue("latitude")
		lonStr := r.FormValue("longitude")

		if title == "" || dateStr == "" || timeStr == "" {
			http.Error(w, "Le titre, la date et l'heure sont obligatoires", http.StatusBadRequest)
			return
		}

		if address == "" && (latStr == "" || lonStr == "") {
			http.Error(w, "Veuillez renseigner une adresse ou un point sur la carte", http.StatusBadRequest)
			return
		}

		// ---------------------------
		// 1) PARSE DATE HEURE EN FRANCE
		// ---------------------------
		startStr := dateStr + " " + timeStr
		locParis, _ := time.LoadLocation("Europe/Paris")

		// Parse en fuseau France (IMPORTANT)
		startDatetimeParis, err := time.ParseInLocation("2006-01-02 15:04", startStr, locParis)
		if err != nil {
			log.Printf("ERREUR parsing datetime France : %v", err)
			http.Error(w, "Date ou heure invalide", http.StatusBadRequest)
			return
		}

		// Conversion UTC pour stockage
		startDatetimeUTC := startDatetimeParis.UTC()

		// ---------------------------
		// RÉCUPÉRATION ADRESSE + COORDS
		// ---------------------------
		var lat, lon float64
		var finalAddress string

		// Cas 1 : coordonnées présentes → clique carte
		if latStr != "" && lonStr != "" {
			lat, err = strconv.ParseFloat(latStr, 64)
			if err != nil {
				http.Error(w, "Latitude invalide", http.StatusBadRequest)
				return
			}

			lon, err = strconv.ParseFloat(lonStr, 64)
			if err != nil {
				http.Error(w, "Longitude invalide", http.StatusBadRequest)
				return
			}

			if address != "" {
				finalAddress = address
			} else {
				finalAddress = data.Address
			}

			// Cas 2 : adresse manuelle
		} else if address != "" {
			lat, lon, err = utils.GeocodeAddress(address)
			if err != nil {
				http.Error(w, "Adresse introuvable", http.StatusBadRequest)
				return
			}
			finalAddress = address

			// Cas 3 : rien fourni → coordonnées user
		} else {
			lat = data.Latitude
			lon = data.Longitude
			finalAddress = data.Address
		}

		// ---------------------------
		// INSERT DATABASE (UTC)
		// ---------------------------
		_, err = db.Exec(`
            INSERT INTO events (title, created_by, latitude, longitude, address, start_datetime)
            VALUES (?, ?, ?, ?, ?, ?)
        `, title, session.UserID, lat, lon, finalAddress, startDatetimeUTC)

		if err != nil {
			log.Printf("ERREUR insertion event: %v", err)
			http.Error(w, "Impossible de créer l'événement", http.StatusInternalServerError)
			return
		}

		log.Printf("✅ Événement créé avec succès: %s (%s)", title, startDatetimeParis.Format("15:04"))

		http.Redirect(w, r, "/RideUp", http.StatusSeeOther)
		return
	}

	// ---------------------------
	// AFFICHAGE FORMULAIRE
	// ---------------------------
	err = NewEventHtml.Execute(w, data)
	if err != nil {
		log.Printf("Erreur template NewEventHtml: %v", err)
		utils.NotFoundHandler(w)
	}
}
