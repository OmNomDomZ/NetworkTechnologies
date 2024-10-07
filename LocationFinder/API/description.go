package API

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
)

type Description struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func GetDescription(id int, ch chan<- Description) {
	req, err := http.NewRequest("GET", "https://kudago.com/public-api/v1.4/places/"+strconv.Itoa(id), nil)
	if err != nil {
		fmt.Println(err)
		ch <- Description{}
		return
	}

	q := req.URL.Query()
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		ch <- Description{}
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		ch <- Description{}
		return
	}

	var description Description
	err = json.Unmarshal(body, &description)
	if err != nil {
		fmt.Println(err)
		ch <- Description{}
		return
	}

	description.Description = removeHTMLTags(description.Description)

	ch <- description
}

func removeHTMLTags(input string) string {
	re := regexp.MustCompile(`<.*?>`)
	return re.ReplaceAllString(input, "")
}
