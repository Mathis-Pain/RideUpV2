package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type contextKey string

const UserIDKey contextKey = "user_id"

var jwksURL string
var keySet jwk.Set

func Init(supabaseURL string) error {
	jwksURL = supabaseURL + "/auth/v1/.well-known/jwks.json"
	ks, err := jwk.Fetch(context.Background(), jwksURL)
	if err != nil {
		return err
	}
	keySet = ks
	return nil
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			http.Error(w, `{"error":"missing token"}`, http.StatusUnauthorized)
			return
		}

		parsed, err := jwt.Parse([]byte(token), jwt.WithKeySet(keySet))
		if err != nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		sub, ok := parsed.Get("sub")
		if !ok {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, sub.(string))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractToken(r *http.Request) string {
	// WS upgrade: ?token=...
	if t := r.URL.Query().Get("token"); t != "" {
		return t
	}
	// REST: Authorization: Bearer ...
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	return ""
}
