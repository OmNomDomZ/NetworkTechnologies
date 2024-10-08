package Handlers

import (
	"html/template"
	"locations/API"
	"net/http"
	"strconv"
	"sync"
)

var tmplLocInf = template.Must(template.ParseFiles("html/locationInformation.html"))

func LocationInformationHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	lat := r.URL.Query().Get("lat")
	lng := r.URL.Query().Get("lng")

	if name == "" {
		http.Error(w, "Место не указано", http.StatusBadRequest)
		return
	}

	if lat == "" || lng == "" {
		http.Error(w, "Координаты не указаны", http.StatusBadRequest)
		return
	}

	latVal, _ := strconv.ParseFloat(lat, 64)
	lngVal, _ := strconv.ParseFloat(lng, 64)

	weatherChan := make(chan API.Weather, 1)
	placesChan := make(chan API.Places, 1)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		API.GetPlaces(API.Point{Lat: latVal, Lng: lngVal}, placesChan)
	}()

	go func() {
		defer wg.Done()
		API.GetWeather(API.Point{Lat: latVal, Lng: lngVal}, weatherChan)
	}()

	go func() {
		wg.Wait()
		close(weatherChan)
		close(placesChan)
	}()

	weather := <-weatherChan
	places := <-placesChan

	var descWG sync.WaitGroup
	descriptionChan := make(chan API.Description)
	var placesDescriptions []API.Description

	for _, place := range places.Results {
		descWG.Add(1)
		go func(placeId int) {
			defer descWG.Done()
			API.GetDescription(placeId, descriptionChan)
		}(place.Id)
	}

	go func() {
		descWG.Wait()
		close(descriptionChan)
	}()

	for place := range descriptionChan {
		placesDescriptions = append(placesDescriptions, place)
	}

	data := struct {
		Name               string
		Lat                float64
		Lng                float64
		Description        string
		Temp               float64
		Results            []API.Place
		PlacesDescriptions []API.Description
	}{
		Name:               name,
		Lat:                latVal,
		Lng:                lngVal,
		Description:        weather.Weather[0].Description,
		Temp:               weather.Main.Temp,
		Results:            places.Results,
		PlacesDescriptions: placesDescriptions,
	}

	err := tmplLocInf.ExecuteTemplate(w, "locationInformation.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
