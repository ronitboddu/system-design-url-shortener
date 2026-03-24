package main

import (
	"bytes"
	"fmt"
	"net/http"
)

type Post struct {
	ExpTime int    `json:"expTime"`
	UrlPath string `json:"urlPath"`
}

func main() {
	postUrl := "http://localhost:8080/shorten"
	body := []byte(`{
		"expTime" : 2,
		"urlPath" : "https://example.com/watch-movie/the-lord-of-the-rings"
	}`)

	r, err := http.NewRequest("POST", postUrl, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}

	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(r)

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	fmt.Println("response Status:", res.Status)
}
