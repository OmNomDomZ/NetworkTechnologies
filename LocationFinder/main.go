package main

import (
	"bufio"
	"fmt"
	"locations/API"
	"os"
	"strconv"
	"strings"
	"sync"
)

func main() {
	loc := strings.Join(os.Args[1:], " ")

	locations := API.GetLocation(loc)
	if len(locations.Hits) == 0 {
		fmt.Println("No results found")
		return
	}

	for i, location := range locations.Hits {
		fmt.Printf("%v. Локация: %s, Страна: %s, Город: %s, Координаты %f, %f\n",
			i, location.Name, location.Country, location.City, location.Point.Lat, location.Point.Lng)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Введите номер выбранной локации: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	index, err := strconv.Atoi(input)
	if err != nil || index >= len(locations.Hits) || index < 0 {
		fmt.Println("Неверный выбор")
		return
	}

	selectedLocation := locations.Hits[index]
	fmt.Printf("Вы выбрали: %s, %s, %s. Координаты: %f, %f\n\n",
		selectedLocation.Name, selectedLocation.Country, selectedLocation.City, selectedLocation.Point.Lat, selectedLocation.Point.Lng)

	weatherChan := make(chan API.Weather, 1)
	placesChan := make(chan API.Places, 1)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		API.GetWeather(selectedLocation.Point, weatherChan)
	}()

	go func() {
		defer wg.Done()
		API.GetPlaces(selectedLocation.Point, placesChan)
	}()

	go func() {
		wg.Wait()
		close(weatherChan)
		close(placesChan)
	}()

	weather := <-weatherChan
	places := <-placesChan

	fmt.Printf("Погода: %s, Температура: %.2f\n", weather.Weather[0].Description, weather.Main.Temp)

	for i, place := range places.Results {
		fmt.Printf("%v. id: %v, Название: %s, Сайт: %s\n",
			i, place.Id, place.Title, place.SiteURL)
	}

	var descWG sync.WaitGroup
	descriptionChan := make(chan API.Description)

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

	for description := range descriptionChan {
		fmt.Printf("Название: %s\nОписание: %s\n\n", description.Title, description.Description)
	}
}
