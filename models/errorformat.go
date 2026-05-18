package models

// Erreurs dans le formulaire d'inscription
type RegisterDataError struct {
	NameError  string
	EmailError string
	PassError  string
}
type User struct {
	ID        int
	Username  string
	Password  string
	Email     string
	ProfilPic string
	Status    string
}

type UpdatePassword struct {
	OldPasswordError        string
	NewPasswordError        string
	ConfirmNewPasswordError string
}
