package handlers

import (
	"RideUP/utils"
	"html/template"
	"log"
	"net/http"
)

var HomeHtml = template.Must(template.ParseFiles("templates/home.html", "templates/inithtml/inithead.html", "templates/inithtml/initfooter.html"))

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	err := HomeHtml.Execute(w, nil)
	if err != nil {
		log.Printf("Erreur lors de l'exécution du template HomeHtml: %v", err)
		utils.NotFoundHandler(w)
	}
}
