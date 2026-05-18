package middleware

import (
	"RideUP/sessions"
	"net/http"
)

// AuthMiddleware protège les routes nécessitant une session
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie == nil {
			http.Redirect(w, r, "/registration", http.StatusFound)
			return
		}

		session, err := sessions.GetSession(cookie.Value)
		if err != nil || session.UserID == 0 {
			http.Redirect(w, r, "/registration", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
