package API

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
	Point   Point  `json:"point"`
	Name    string `json:"name"`
	Country string `json:"country"`
	City    string `json:"city"`
}

type GraphhopperAnswer struct {
	Hits []Location `json:"hits"`
}

func GetLocation(location string) GraphhopperAnswer {
	req, err := http.NewRequest("GET", "https://graphhopper.com/api/1/geocode", nil)
	if err != nil {
		fmt.Println(err)
	}

	q := req.URL.Query()
	q.Add("q", location)
	q.Add("key", "a98f475e-eaa3-4145-b8be-67aee6cdec09")
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

	return gha
}
