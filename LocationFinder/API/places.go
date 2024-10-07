package API

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Place struct {
	Id      int    `json:"id"`
	Title   string `json:"title"`
	SiteURL string `json:"site_url"`
}

type Places struct {
	Results []Place `json:"results"`
}

func GetPlaces(point Point, ch chan<- Places) {
	req, err := http.NewRequest("GET", "https://kudago.com/public-api/v1.4/places", nil)
	if err != nil {
		fmt.Println(err)
		ch <- Places{}
		return
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
		ch <- Places{}
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		ch <- Places{}
		return
	}

	var places Places
	err = json.Unmarshal(body, &places)
	if err != nil {
		fmt.Println(err)
		ch <- Places{}
		return
	}

	ch <- places
}
