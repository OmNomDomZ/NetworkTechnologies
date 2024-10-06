package API

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Places struct {
	Results []struct {
		Id      int    `json:"id"`
		Title   string `json:"title"`
		SiteURL string `json:"site_url"`
	} `json:"results"`
}

func GetPlaces(point Point) Places {
	req, err := http.NewRequest("GET", "https://kudago.com/public-api/v1.4/places", nil)
	if err != nil {
		fmt.Println(err)
	}

	q := req.URL.Query()
	q.Add("lat", fmt.Sprintf("%f", point.Lat))
	q.Add("lon", fmt.Sprintf("%f", point.Lng))
	q.Add("radius", "10000")
	q.Add("page_size", "5")
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

	var places Places
	err = json.Unmarshal(body, &places)
	if err != nil {
		fmt.Println(err)
	}

	return places
}
