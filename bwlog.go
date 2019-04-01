package main

import (
	_ "./statik" // TODO: Replace with the absolute import path
	"fmt"
	"github.com/rakyll/statik/fs"
	"log"
	"net/http"
	"strings"
	"flag"
)

type Config struct {
	Interfaces []string
	Database   string
	Save       int
	Listen     string
}

func main() {
	interfaces := flag.String("i", "eth0", "interfaces to monitor, comma separated")
	listen := flag.String("l", ":8080", "port to listen on")
	database := flag.String("d", "./bwlog.sqlite", "database path")
	save := flag.Int("s", 60, "save to database every X seconds")

	flag.Parse()

	var config Config
	config.Interfaces = strings.Split(*interfaces, ",")
	config.Database = *database
	config.Listen = *listen
	config.Save = *save

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
