package handlers

import (
	"RideUP/sessions"
	"RideUP/utils"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

func JoinEventHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(">>> JoinEventHandler appelé")

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	session, err := sessions.GetSessionFromRequest(r)
	if err != nil {
		log.Println(" Utilisateur non connecté")
		http.Error(w, "Non connecté", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Printf(" Erreur ParseForm: %v", err)
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	eventID := r.FormValue("event_id")
	action := r.FormValue("action")

	log.Printf("📝 EventID: %s, Action: %s, UserID: %d", eventID, action, session.UserID)

	if eventID == "" || (action != "join" && action != "leave") {
		log.Println(" Paramètres invalides")
		http.Error(w, "Paramètres invalides", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", "./data/RideUp.db")
	if err != nil {
		log.Printf(" Erreur ouverture DB: %v", err)
		utils.InternalServError(w)
		return
	}
	defer db.Close()

	var participants int64
	var joined bool

	if action == "join" {
		// Vérifier si déjà inscrit
		var count int
		err = db.QueryRow(`SELECT COUNT(*) FROM event_participants WHERE user_id = ? AND event_id = ?`,
			session.UserID, eventID).Scan(&count)
		if err != nil {
			log.Printf(" Erreur vérification inscription: %v", err)
			http.Error(w, "Erreur base de données", http.StatusInternalServerError)
			return
		}

		if count == 0 {
			// Inscrire l'utilisateur
			_, err = db.Exec(`INSERT INTO event_participants (user_id, event_id) VALUES (?, ?)`,
				session.UserID, eventID)
			if err != nil {
				log.Printf("Erreur insertion participant: %v", err)
				http.Error(w, "Erreur insertion", http.StatusInternalServerError)
				return
			}

			// Incrémenter le compteur
			_, err = db.Exec(`UPDATE events SET participants = COALESCE(participants, 0) + 1 WHERE id = ?`, eventID)
			if err != nil {
				log.Printf("Erreur mise à jour compteur: %v", err)
			}

			joined = true
			log.Printf(" Utilisateur %d inscrit à l'événement %s", session.UserID, eventID)
		} else {
			joined = true // Déjà inscrit
			log.Printf("Utilisateur %d déjà inscrit à l'événement %s", session.UserID, eventID)
		}

	} else if action == "leave" {
		// Désinscrire l'utilisateur
		result, err := db.Exec(`DELETE FROM event_participants WHERE user_id = ? AND event_id = ?`,
			session.UserID, eventID)
		if err != nil {
			log.Printf(" Erreur suppression participant: %v", err)
			http.Error(w, "Erreur suppression", http.StatusInternalServerError)
			return
		}

		// Vérifier si une ligne a été supprimée
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			// Décrémenter le compteur
			_, err = db.Exec(`UPDATE events SET participants = MAX(0, COALESCE(participants, 0) - 1) WHERE id = ?`, eventID)
			if err != nil {
				log.Printf("Erreur mise à jour compteur: %v", err)
			}
			log.Printf("Utilisateur %d désinscrit de l'événement %s", session.UserID, eventID)
		}

		joined = false
	}

	// Récupérer le nombre de participants
	err = db.QueryRow(`SELECT COALESCE(participants, 0) FROM events WHERE id = ?`, eventID).Scan(&participants)
	if err != nil {
		log.Printf(" Erreur récupération participants: %v", err)
		participants = 0
	}

	response := map[string]interface{}{
		"joined":       joined,
		"participants": participants,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Erreur encodage JSON: %v", err)
	}

	log.Printf("✅ Réponse envoyée: joined=%v, participants=%d", joined, participants)
}
