package API

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Weather struct {
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temp float64 `json:"temp"`
	}
}

func GetWeather(point Point) {
	req, err := http.NewRequest("GET", "https://api.openweathermap.org/data/2.5/weather", nil)
	if err != nil {
		fmt.Println(err)
	}

	q := req.URL.Query()
	q.Add("appid", "")
	q.Add("lat", fmt.Sprintf("%f", point.Lat))
	q.Add("lon", fmt.Sprintf("%f", point.Lng))
	q.Add("units", "metric")
	q.Add("lang", "ru")
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

	var weather Weather
	err = json.Unmarshal(body, &weather)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Погода: %s, Температура: %.2f\n", weather.Weather[0].Description, weather.Main.Temp)
}
