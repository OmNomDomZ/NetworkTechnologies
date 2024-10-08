package Handlers

import (
	"html/template"
	"locations/API"
	"net/http"
)

var tmplLoc = template.Must(template.ParseFiles("html/locations.html"))

func LocationHandler(w http.ResponseWriter, r *http.Request) {
	loc := r.URL.Query().Get("location")

	if loc == "" {
		http.Error(w, "Локация не указана", http.StatusBadRequest)
		return
	}

	locations := API.GetLocation(loc)
	if len(locations.Hits) == 0 {
		http.Error(w, "Локация не найдена", http.StatusBadRequest)
		return
	}

	err := tmplLoc.ExecuteTemplate(w, "locations.html", locations)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
