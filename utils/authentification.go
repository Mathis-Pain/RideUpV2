package utils

import (
	"RideUP/models"
	"RideUP/utils/getdata"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func Authentification(db *sql.DB, email string, password string) (models.User, error) {
	if email == "" || password == "" {
		mylog := fmt.Errorf("tous les champs sont requis")
		log.Println("Erreur : <authentification.go>", mylog)
		return models.User{}, mylog
	}
	// Récupère l'Email et le mot de passe (crypté) à partir de l'identifiant
	user, err := getdata.GetUserFromLogin(db, email)
	if errors.Is(err, sql.ErrNoRows) {
		// Si aucun utilisateur n'est trouvé avec cet identifiant (mail ou pseudo), renvoie une erreur
		log.Printf("ERREUR : <authentification.go> Tentative de connexion échouée : L'utilisateur %s n'existe pas.\n", email)
		return models.User{}, fmt.Errorf("nom d'utilisateur incorrect")
	} else if err != nil {
		// Erreur dans la base de données
		mylog := fmt.Errorf("(db) Impossible de récupérer les données utilisateur dans la base de données : %v", err)
		log.Println("ERREUR : <authentification.go> ", mylog)
		return models.User{}, mylog
	}

	// Fonction bcrypt pour comparer le mot de passe entré par l'utilisateur avec celui présent dans la base de données
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Println("ERREUR : <authentification.go> : Mot de passe incorrect")
		return models.User{}, fmt.Errorf("mot de passe incorrect")
	}
	return user, err
}
