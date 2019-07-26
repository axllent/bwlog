package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "./statik"
	"github.com/axllent/gitrel"
	"github.com/rakyll/statik/fs"
)

// Config struct
type Config struct {
	Interfaces []string
	Database   string
	Save       int
	Listen     string
}

var version = "dev"

func main() {
	interfaces := flag.String("i", "", "interfaces to monitor, comma separated eg: eth0,eth1")
	listen := flag.String("l", "0.0.0.0:8080", "port to listen on")
	database := flag.String("d", "", "database directory path")
	save := flag.Int("s", 60, "save to database every X seconds")
	update := flag.Bool("u", false, "update to latest release")
	showversion := flag.Bool("v", false, "show version number")

	flag.Usage = func() {
		fmt.Println(fmt.Sprintf("BWLog %s: A lightweight bandwidth logger.\n", version))
		fmt.Println(fmt.Sprintf("Usage example: %s -i eth0 -l 0.0.0.0:8080 -d ~/bwlog/\n", os.Args[0]))
		fmt.Println("Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	var config Config
	config.Interfaces = strings.Split(*interfaces, ",")
	config.Database = *database
	config.Listen = *listen
	config.Save = *save

	if *showversion {
		fmt.Println(fmt.Sprintf("Version: %s", version))
		latest, _, _, err := gitrel.Latest("axllent/bwlog", "bwlog")
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

	if *interfaces == "" {
		PrintErr("No network interfaces specified.\n")
		fmt.Println(fmt.Sprintf("Usage example: %s -i eth0 -l 0.0.0.0:8080 -d ~/bwlog/\n", os.Args[0]))
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *database == "" {
		PrintErr("No database directory specified.\n")
		fmt.Println(fmt.Sprintf("Usage example: %s -i eth0 -l 0.0.0.0:8080 -d ~/bwlog/\n", os.Args[0]))
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	dbinfo, err := os.Stat(config.Database)
	if err != nil {
		PrintErr(fmt.Sprintf("%s does not exist, exiting", config.Database))
		os.Exit(1)
	}

	if !dbinfo.IsDir() {
		PrintErr(fmt.Sprintf("%s is not a directory, exiting", config.Database))
		os.Exit(1)
	}

	// Start new thread for httpd
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

		PrintInfo(fmt.Sprintf("HTTP listening on %s", config.Listen))

		log.Fatal(http.ListenAndServe(config.Listen, nil))
	}()

	// Stats daemon
	bwLogger(config)
}
