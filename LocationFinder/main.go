package main

import (
	"bufio"
	"fmt"
	"locations/API"
	"os"
	"strconv"
	"strings"
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
	if err != nil || index > len(locations.Hits) || index < 0 {
		fmt.Println("Неверный выбор")
		return
	}

	selectedLocation := locations.Hits[index]
	fmt.Printf("Вы выбрали: %s, %s, %s. Координаты: %f, %f\n\n",
		selectedLocation.Name, selectedLocation.Country, selectedLocation.City, selectedLocation.Point.Lat, selectedLocation.Point.Lng)

	API.GetWeather(selectedLocation.Point)
	places := API.GetPlaces(selectedLocation.Point)

	for i, place := range places.Results {
		fmt.Printf("%v. id: %v, Название: %s, Сайт: %s\n",
			i, place.Id, place.Title, place.SiteURL)
	}

	for _, place := range places.Results {
		description := API.GetDescription(place.Id)
		fmt.Printf("Название: %s\nОписание: %s\n\n", description.Title, description.Description)
	}
}
