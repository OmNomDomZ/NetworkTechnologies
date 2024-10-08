package Handlers

import (
	"html/template"
	"net/http"
)

var tmplSearch = template.Must(template.ParseFiles("html/search.html"))

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	err := tmplSearch.ExecuteTemplate(w, "search.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
