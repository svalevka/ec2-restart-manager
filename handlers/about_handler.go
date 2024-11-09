package handlers

import (
	"html/template"
	"log"
	"net/http"
)

var aboutTemplate = template.Must(template.ParseFiles("templates/layout.html", "templates/about.html"))

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title      string
		IsLoggedIn bool
	}{
		Title:      "About EC2 Manager",
		IsLoggedIn: false,
	}

	if err := aboutTemplate.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Error rendering about template: %v", err)
	}
}
