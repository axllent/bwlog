//go:generate bin/statik -f -src=./web

package main

import (
	_ "./statik"
	"flag"
	"fmt"
	"github.com/axllent/gitrel"
	"github.com/rakyll/statik/fs"
	"log"
	"net/http"
	"strings"
	"os"
)

type Config struct {
	Interfaces []string
	Database   string
	Save       int
	Listen     string
}

var version = "dev"

func main() {
	interfaces := flag.String("i", "eth0", "interfaces to monitor, comma separated")
	listen := flag.String("l", "0.0.0.0:8080", "port to listen on")
	database := flag.String("d", "./bwlog.sqlite", "database path")
	save := flag.Int("s", 60, "save to database every X seconds")
	update := flag.Bool("u", false, "updater to latest release")
	showversion := flag.Bool("v", false, "show version number")

	flag.Parse()

	var config Config
	config.Interfaces = strings.Split(*interfaces, ",")
	config.Database = *database
	config.Listen = *listen
	config.Save = *save

	if *showversion {
		fmt.Println(fmt.Sprintf("Version: %s", version))
		latest, _, _, err := gitrel.Latest("axllent/bwlog", "bwlog");
		if err == nil && latest != version {
			fmt.Println(fmt.Sprintf("Update available: %s\nRun `%s -u` to update.", latest, os.Args[0]))
		}
		return
	}

	if *update {
		rel, err := gitrel.Update("axllent/bwlog", "bwlog", version)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(fmt.Sprintf("Updated %s to version %s", os.Args[0], rel))
		return
	}

	go func() {
		// load static file FS
		statikFS, err := fs.New()
		if err != nil {
			log.Fatal(err)
		}

		// default http route (statik FS)
		http.Handle("/", http.FileServer(statikFS))

		// stats controller
		http.HandleFunc("/stats/", func(w http.ResponseWriter, r *http.Request) {
			statsController(w, r, config)
		})

		// websocket route
		http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
			streamController(w, r, config)
		})

		log.Println(fmt.Sprintf("HTTP listening on %s", config.Listen))

		log.Fatal(http.ListenAndServe(config.Listen, nil))
	}()

	// Stats daemon
	bwLogger(config)
}
