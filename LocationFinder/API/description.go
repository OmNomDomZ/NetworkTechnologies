package API

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Description struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func GetDescription(id int) Description {
	req, err := http.NewRequest("GET", "https://kudago.com/public-api/v1.4/places/"+strconv.Itoa(id), nil)
	if err != nil {
		fmt.Println(err)
	}

	q := req.URL.Query()
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

	var description Description
	err = json.Unmarshal(body, &description)
	if err != nil {
		fmt.Println(err)
	}

	return description
}
