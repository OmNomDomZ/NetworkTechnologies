package main

import (
	"fmt"
	"locations/Handlers"
	"net/http"
)

func main() {
	http.HandleFunc("/", Handlers.HomeHandler)
	http.HandleFunc("/search", Handlers.SearchHandler)
	http.HandleFunc("/locations", Handlers.LocationHandler)
	http.HandleFunc("/locationIformation", Handlers.LocationInformationHandler)

	fmt.Println("Запуск сервера на порту 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}
