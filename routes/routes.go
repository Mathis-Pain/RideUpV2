package routes

import (
	"RideUP/handlers"
	"RideUP/middleware"
	"RideUP/utils"
	"RideUP/utils/authextern"
	"net/http"
)

func InitRoutes() *http.ServeMux {

	mux := http.NewServeMux()

	// route Home
	// fonction anonyme utilisé pour faire dautre verification
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			utils.NotFoundHandler(w)
			return
		}
		handlers.HomeHandler(w, r)
	})
	mux.HandleFunc("/CreateAccount", handlers.RegisterHandler)
	mux.HandleFunc("/Connect", handlers.ConnectHandler)
	mux.HandleFunc("/Disconnect", handlers.DisconnectHandler)
	mux.Handle("/RideUp", middleware.AuthMiddleware(http.HandlerFunc(handlers.RideUpHandler)))
	mux.Handle("/NewEvent", middleware.AuthMiddleware(http.HandlerFunc(handlers.NewEventHandler)))
	mux.HandleFunc("/JoinEvent", handlers.JoinEventHandler)
	mux.Handle("/Preference", middleware.AuthMiddleware(http.HandlerFunc(handlers.PreferenceHandler)))
	mux.Handle("/UpdatePassword", middleware.AuthMiddleware(http.HandlerFunc(handlers.UpdatePasswordHandler)))
	mux.Handle("/Admin", middleware.AuthMiddleware(middleware.AdminOnly(http.HandlerFunc(handlers.RideUpAdminHandler))))
	// charger par le js dans navbar
	mux.HandleFunc("/api/check-admin", handlers.CheckAdminHandler)

	// Authentification par google ou github
	mux.HandleFunc("/auth/google/login", authextern.HandleGoogleLogin)
	mux.HandleFunc("/auth/google/callback", authextern.HandleGoogleCallback)
	// servir les fichiers static
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	return mux
}
