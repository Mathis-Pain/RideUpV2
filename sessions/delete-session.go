package sessions

import (
	"database/sql"
	"log"
)

// Supprime une session
func DeleteSession(sessionID string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	return err
}

// CleanupExpiredSessions supprime les sessions expirées
func CleanupExpiredSessions() error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	res, err := db.Exec(`
		DELETE FROM sessions
		WHERE datetime(expires_at) < datetime('now')
	`)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	log.Printf("🧹 Sessions expirées supprimées : %d\n", count)
	return nil

}
