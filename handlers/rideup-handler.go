package handlers

import (
	"RideUP/models"
	"RideUP/sessions"
	"RideUP/utils"
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var EventHtml = template.Must(template.ParseFiles(
	"templates/rideup.html",
	"templates/inithtml/inithead.html",
	"templates/inithtml/initnav.html",
	"templates/inithtml/initfooter.html",
))

func RideUpHandler(w http.ResponseWriter, r *http.Request) {

	// Session
	session, err := sessions.GetSessionFromRequest(r)
	if err != nil {
		http.Redirect(w, r, "/Connect", http.StatusSeeOther)
		return
	}
	userID := session.UserID

	//  DB
	db, err := sql.Open("sqlite3", "./data/RideUp.db")
	if err != nil {
		utils.InternalServError(w)
		return
	}
	defer db.Close()

	// Suppression événements passés
	db.Exec(`DELETE FROM events WHERE datetime(start_datetime) < datetime('now')`)

	// Suppression manuelle envoyé par rideup.js via le post js
	if r.Method == http.MethodPost {
		eventID := r.FormValue("event_id")
		action := r.FormValue("action")

		// securisation du delete via userID
		if action == "delete" {
			db.Exec(`DELETE FROM events WHERE id = ? AND created_by = ?`, eventID, userID)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"success": true}`))
			return
		}
	}

	// ---------------------------
	//  RÉCUPÉRER POSITION + PRÉFÉRENCES
	// ---------------------------

	var userLat, userLon float64
	var pref int

	err = db.QueryRow(`
        SELECT latitude, longitude, preference 
        FROM users 
        WHERE id = ?`, userID).Scan(&userLat, &userLon, &pref)

	if err != nil {
		userLat = 48.8566
		userLon = 2.3522
		pref = 50
	}

	//  Fuseau France
	locParis, _ := time.LoadLocation("Europe/Paris")

	// ---------------------------
	//  ÉVÉNEMENTS DE L'UTILISATEUR
	// ---------------------------
	// ASC tri ascendant "e" alias
	userRows, err := db.Query(`
        SELECT e.id, e.title, e.description, e.created_by, u.username, e.created_at,
               e.latitude, e.longitude, e.address, e.start_datetime, e.end_datetime, e.participants
        FROM events e
        JOIN users u ON e.created_by = u.id
        WHERE e.created_by = ?
        ORDER BY e.start_datetime ASC`, userID)
	if err != nil {
		log.Printf("❌ Erreur query userEvents: %v", err)
		utils.InternalServError(w)
		return
	}
	defer userRows.Close()

	var userEvents []models.Event

	for userRows.Next() {
		var e models.Event
		var endDatetimeStr sql.NullString
		var participantsNull sql.NullInt64

		err := userRows.Scan(
			&e.ID,
			&e.Title,
			&e.Description,
			&e.CreatedBy,
			&e.CreatorName,
			&e.CreatedAt,
			&e.Latitude,
			&e.Longitude,
			&e.Address,
			&e.StartDatetime,
			&endDatetimeStr,
			&participantsNull,
		)
		if err != nil {
			log.Printf("❌ Erreur scan userEvent: %v", err)
			continue
		}

		//  Convertir participants en int
		if participantsNull.Valid {
			e.Participants = int(participantsNull.Int64)
		} else {
			e.Participants = 0
		}

		//  Remplir DescriptionStr pour JSON
		if e.Description.Valid {
			e.DescriptionStr = e.Description.String
		} else {
			e.DescriptionStr = ""
		}

		//  Conversion end_datetime
		if endDatetimeStr.Valid {
			endTime, err := time.Parse("2006-01-02 15:04:05", endDatetimeStr.String)
			if err == nil {
				e.EndDatetime = &endTime
			}
		}

		//  Conversion France pour HTML
		e.StartDatetime = e.StartDatetime.In(locParis)
		e.FormattedStart = e.StartDatetime.Format("le 02/01/2006 à 15:04")

		if e.EndDatetime != nil {
			t := e.EndDatetime.In(locParis)
			e.EndDatetime = &t
		}

		//  Vérifier si rejoint (créateur auto-rejoint)
		var count int
		db.QueryRow(`SELECT COUNT(*) FROM event_participants WHERE user_id = ? AND event_id = ?`,
			userID, e.ID).Scan(&count)
		e.UserJoined = count > 0 || e.CreatedBy == userID // ✅ AJOUT ICI

		userEvents = append(userEvents, e)
	}

	log.Printf("✅ Événements utilisateur: %d", len(userEvents))

	// ---------------------------
	//  TOUS LES ÉVÉNEMENTS
	// ---------------------------
	allRows, err := db.Query(`
        SELECT e.id, e.title, e.description, e.created_by, u.username, e.created_at,
               e.latitude, e.longitude, e.address, e.start_datetime, e.end_datetime, e.participants
        FROM events e
        JOIN users u ON e.created_by = u.id
        ORDER BY e.start_datetime ASC`)
	if err != nil {
		log.Printf("❌ Erreur query allEvents: %v", err)
		utils.InternalServError(w)
		return
	}
	defer allRows.Close()

	var availableEvents []models.Event
	var availableEventsForMap []models.Event

	for allRows.Next() {
		var e models.Event
		var endDatetimeStr sql.NullString
		var participantsNull sql.NullInt64

		err := allRows.Scan(
			&e.ID,
			&e.Title,
			&e.Description,
			&e.CreatedBy,
			&e.CreatorName,
			&e.CreatedAt,
			&e.Latitude,
			&e.Longitude,
			&e.Address,
			&e.StartDatetime,
			&endDatetimeStr,
			&participantsNull,
		)
		if err != nil {
			log.Printf("❌ Erreur scan allEvent: %v", err)
			continue
		}

		//  Convertir participants en int
		if participantsNull.Valid {
			e.Participants = int(participantsNull.Int64)
		} else {
			e.Participants = 0
		}

		//  Remplir DescriptionStr pour JSON
		if e.Description.Valid {
			e.DescriptionStr = e.Description.String
		} else {
			e.DescriptionStr = ""
		}

		//  Conversion end_datetime
		if endDatetimeStr.Valid {
			endTime, err := time.Parse("2006-01-02 15:04:05", endDatetimeStr.String)
			if err == nil {
				e.EndDatetime = &endTime
			}
		}

		// Vérifier si rejoint (créateur auto-rejoint)
		var count int
		db.QueryRow(`SELECT COUNT(*) FROM event_participants WHERE user_id = ? AND event_id = ?`,
			userID, e.ID).Scan(&count)
		e.UserJoined = count > 0 || e.CreatedBy == userID // ✅ AJOUT ICI

		//  Copie UTC pour JSON de la carte
		eForMap := e
		eForMap.StartDatetime = e.StartDatetime.UTC()
		if eForMap.EndDatetime != nil {
			endUTC := eForMap.EndDatetime.UTC()
			eForMap.EndDatetime = &endUTC
		}
		availableEventsForMap = append(availableEventsForMap, eForMap)

		//  Conversion France pour HTML
		e.StartDatetime = e.StartDatetime.In(locParis)
		e.FormattedStart = e.StartDatetime.Format("le 02/01/2006 à 15:04")

		if e.EndDatetime != nil {
			t := e.EndDatetime.In(locParis)
			e.EndDatetime = &t
		}

		availableEvents = append(availableEvents, e)
	}

	log.Printf("✅ Total événements: %d", len(availableEvents))

	//  Filtrage par préférence
	availableEventsFilter := utils.FilterPreference(availableEvents, userLat, userLon, pref)
	log.Printf("✅ Événements après filtre: %d", len(availableEventsFilter))

	//  JSON avec les données UTC pour la carte
	eventsJSON, err := json.Marshal(availableEventsForMap)
	if err != nil {
		log.Printf("❌ Erreur marshaling JSON: %v", err)
		eventsJSON = []byte("[]")
	}

	log.Printf("📦 JSON généré: %d bytes pour %d events", len(eventsJSON), len(availableEventsForMap))
	if len(eventsJSON) > 0 {
		log.Printf("🔍 Premier event JSON: %s", string(eventsJSON[:min(300, len(eventsJSON))]))
	}

	// ---------------------------
	//  TEMPLATE
	// ---------------------------
	data := struct {
		ActivePage      string
		UserID          int
		CurrentUserID   int
		UserEvents      []models.Event
		AvailableEvents []models.Event
		Latitude        float64
		Longitude       float64
		EventsJSON      string
	}{
		ActivePage:      "RideUp",
		UserID:          userID,
		CurrentUserID:   userID,
		UserEvents:      userEvents,
		AvailableEvents: availableEventsFilter,
		Latitude:        userLat,
		Longitude:       userLon,
		EventsJSON:      string(eventsJSON),
	}

	if err := EventHtml.Execute(w, data); err != nil {
		log.Printf("❌ Erreur template: %v", err)
		utils.InternalServError(w)
		return
	}
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
