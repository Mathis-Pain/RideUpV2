// Dans utils/getdata/user.go ou similaire
package getdata

import "database/sql"

func IsUserAdmin(db *sql.DB, userID int) (bool, error) {
	var roleid int
	var isAdmin bool
	query := `SELECT role_id FROM users WHERE id = ?`
	err := db.QueryRow(query, userID).Scan(&roleid)
	if err != nil {
		return false, err
	}
	if roleid == 1 {
		isAdmin = true
	} else {
		isAdmin = false
	}
	return isAdmin, nil
}
