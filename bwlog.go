package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Config struct {
	Interfaces []string `json:"if"`
	Database   string   `json:"db"`
	Save       int      `json:"save"`
}

func main() {
	jsonFile, err := os.Open("bwlog.json")

	if err != nil {
		fmt.Println(fmt.Errorf("Error: %s", err))
		return
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var config Config

	json.Unmarshal(byteValue, &config)

	go func() {
		http.HandleFunc("/", handler)
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	LogStats(config)

	fmt.Println("ok")
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}
