package sessions

import (
	"fmt"
	"log"
	"time"
)

func IsValidSession(sessionID string) bool {
	session, err := GetSession(sessionID)
	if err != nil {
		fmt.Println("Session non trouvée")
		return false
	}

	// Vérifier que UserID est valide
	if session.UserID == 0 {
		fmt.Println("UserID invalide")
		return false
	}

	// Si ExpiresAt est zéro, c'est souvent signe que GetSession n'a pas parsé la date
	if session.ExpiresAt.IsZero() {
		log.Println("IsValidSession: ExpiresAt is zero -> vérifier GetSession (parsing de expires_at)")
		return false
	}

	// Comparaison : utiliser UTC si vous stockez en UTC
	now := time.Now()
	if session.ExpiresAt.Before(now) {
		log.Printf("IsValidSession: session expirée (expires_at=%v now=%v)\n", session.ExpiresAt, now)
		return false
	}

	return true
}
