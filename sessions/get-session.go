package sessions

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"RideUP/models"
)

// GetSession récupère une session depuis la DB
func GetSession(sessionID string) (models.Session, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return models.Session{}, err
	}
	defer db.Close()

	var session models.Session
	var dataJSON, expiresAtStr, createdAtStr string

	err = db.QueryRow(`
		SELECT id, user_id, data, expires_at, created_at
		FROM sessions
		WHERE id = ?
	`, sessionID).Scan(&session.ID, &session.UserID, &dataJSON, &expiresAtStr, &createdAtStr)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Session{}, errors.New("session non trouvée")
		}
		return models.Session{}, err
	}
	// Conversion manuelle du TEXT en time.Time
	session.ExpiresAt, err = time.Parse("2006-01-02 15:04:05", expiresAtStr)
	if err != nil {
		return models.Session{}, errors.New("format expires_at invalide")
	}

	session.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		return models.Session{}, errors.New("format created_at invalide")
	}
	// Vérificationd'expiration
	if time.Now().After(session.ExpiresAt) {
		return models.Session{}, errors.New("session expired")
	}

	if err := json.Unmarshal([]byte(dataJSON), &session.Data); err != nil {
		log.Print("<get-session.go> Erreur dans la récupération de session :", err)
		return models.Session{}, err
	}

	return session, nil
}

func GetSessionFromRequest(r *http.Request) (models.Session, error) {

	cookie, err := r.Cookie("session_id")
	if err != nil {
		if err == http.ErrNoCookie {
			return models.Session{}, errors.New("aucun cokie trouvé")
		}
		return models.Session{}, err
	}

	sessionID := cookie.Value
	session, err := GetSession(sessionID)
	if err != nil {
		return models.Session{}, err
	}

	return session, nil
}
