package getdata

import (
	"RideUP/models"
	"database/sql"
)

func GetUserFromLogin(db *sql.DB, login string) (models.User, error) {
	// Préparation de la requête SQL : récupérer id, username et password
	sql := `SELECT id, email, password_hash FROM users WHERE email =?`
	row := db.QueryRow(sql, login)
	var user models.User
	// Parcour la base de données en cherchant le username correspondant
	err := row.Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}
