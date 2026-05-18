package sessions

import (
	"RideUP/models"
	"database/sql"
	"encoding/json"
	"log"
)

// chemin d'acces a la db
const dbPath = "./data/RideUp.db"

// SaveSession sauvegarde ou met à jour une session
// Transforme les données Go en text Json pour pouvoir les sotcker dans la db
func SaveSession(db *sql.DB, session models.Session) error {
	dataJSON, err := json.Marshal(session.Data)
	if err != nil {
		return err
	}
	// ✅ Conversion de time.Time vers un format lisible par SQLite
	expiresAtStr := session.ExpiresAt.Format("2006-01-02 15:04:05")
	createdAtStr := session.CreatedAt.Format("2006-01-02 15:04:05")

	_, err = db.Exec(`
		INSERT OR REPLACE INTO sessions
		(id, user_id, data, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, session.ID, session.UserID, string(dataJSON), expiresAtStr, createdAtStr)

	return err
}

// ouvre juste la db pour sauvegarder la session via SaveSession
func SaveSessionToDB(session models.Session) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Vérifie la connexion tout de suite
	if err := db.Ping(); err != nil {
		return err
	}

	// Sauvegarde la session
	err = SaveSession(db, session)
	if err != nil {
		log.Printf("❌ ERREUR : sauvegarde session échouée : %v", err)
		return err
	}

	log.Printf("✅ Session %s sauvegardée pour l'utilisateur %d", session.ID, session.UserID)
	return nil
}
