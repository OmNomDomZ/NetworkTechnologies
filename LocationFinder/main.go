package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Location struct {
	Point Point  `json:"point"`
	Name  string `json:"name"`
}

type GraphhopperAnswer struct {
	Hits []Location `json:"hits"`
}

func getLocation() {
	req, err := http.NewRequest("GET", "https://graphhopper.com/api/1/geocode", nil)
	if err != nil {
		fmt.Println(err)
	}

	q := req.URL.Query()
	q.Add("q", "Цветной проезд")
	q.Add("key", "")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	var gha GraphhopperAnswer
	err = json.Unmarshal(body, &gha)
	if err != nil {
		fmt.Println(err)
	}

	for _, location := range gha.Hits {
		fmt.Printf("Location: %s, Point %f, %f\n", location.Name, location.Point.Lat, location.Point.Lng)
	}
}

func main() {
	getLocation()
}
