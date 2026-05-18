package getdata

import (
	"database/sql"
	"fmt"
)

func GetPasswordHash(db *sql.DB, userID int) (string, error) {
	var passwordHash string

	query := `SELECT password_hash FROM users WHERE id = ?`
	err := db.QueryRow(query, userID).Scan(&passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("aucun utilisateur trouvé avec l'ID %d", userID)
		}
		return "", err
	}

	return passwordHash, nil
}
