package main

import (
	_ "./statik" // TODO: Replace with the absolute import path
	"encoding/json"
	"fmt"
	"github.com/rakyll/statik/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Config struct {
	Interfaces []string `json:"if"`
	Database   string   `json:"db"`
	Save       int      `json:"save"`
	Listen     string   `json:"listen"`
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
		// load static file FS
		statikFS, err := fs.New()
		if err != nil {
			log.Fatal(err)
		}

		// default http route (statik FS)
		http.Handle("/", http.FileServer(statikFS))

		// websocket route
		http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
			streamController(w, r, config)
		})

		fmt.Println(fmt.Sprintf("HTTP running on %s", config.Listen))

		log.Fatal(http.ListenAndServe(config.Listen, nil))
	}()

	// Stats daemon
	bwLogger(config)
}
