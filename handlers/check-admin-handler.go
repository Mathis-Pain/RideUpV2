package handlers

import (
	"RideUP/sessions"
	"RideUP/utils/getdata"
	"database/sql"
	"encoding/json"
	"net/http"
)

// Appeler par le js dans navbar
func CheckAdminHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	session, err := sessions.GetSessionFromRequest(r)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]bool{"isAdmin": false})
		return
	}

	db, err := sql.Open("sqlite3", "./data/RideUp.db")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]bool{"isAdmin": false})
		return
	}
	defer db.Close()

	// Réutilise la même fonction que le middleware !
	isAdmin, err := getdata.IsUserAdmin(db, session.UserID)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]bool{"isAdmin": false})
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"isAdmin": isAdmin})
}
