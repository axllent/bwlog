//go:generate bin/statik -f -src=./web

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "./statik"
	"github.com/NYTimes/gziphandler"
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

var (
	version        = "dev"
	bauser, bapass string
)

func main() {
	var config Config

	var bauth, interfaces string
	var update, showversion bool
	flag.StringVar(&bauth, "p", "", "basic auth password file (must contain a single <user>:<pass>)")
	flag.StringVar(&interfaces, "i", "", "interfaces to monitor, comma separated eg: eth0,eth1")
	flag.StringVar(&config.Listen, "l", "0.0.0.0:8080", "port to listen on")
	flag.StringVar(&config.Database, "d", "", "database directory path")
	flag.IntVar(&config.Save, "s", 60, "save to database every X seconds")
	flag.BoolVar(&update, "u", false, "update to latest release")
	flag.BoolVar(&showversion, "v", false, "show version number")

	flag.Usage = func() {
		fmt.Println(fmt.Sprintf("BWLog %s: A lightweight bandwidth logger.\n", version))
		fmt.Println(fmt.Sprintf("Usage example: %s -i eth0 -l 0.0.0.0:8080 -d ~/bwlog/\n", os.Args[0]))
		fmt.Println("Options:")
		flag.PrintDefaults()
	}

	flag.Usage = func() {
		fmt.Println(fmt.Sprintf("BWLog %s: A lightweight bandwidth logger.\n", version))
		fmt.Println(fmt.Sprintf("Usage example: %s -i eth0 -l 0.0.0.0:8080 -d ~/bwlog.sqlite\n", os.Args[0]))
		fmt.Println("Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	config.Interfaces = strings.Split(interfaces, ",")

	if showversion {
		fmt.Println(fmt.Sprintf("Version: %s", version))
		latest, _, _, err := gitrel.Latest("axllent/bwlog", "bwlog")
		if err == nil && latest != version {
			fmt.Println(fmt.Sprintf("Update available: %s\nRun `%s -u` to update.", latest, os.Args[0]))
		}
		return
	}

	if update {
		rel, err := gitrel.Update("axllent/bwlog", "bwlog", version)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(fmt.Sprintf("Updated %s to version %s", os.Args[0], rel))
		return
	}

	if interfaces == "" {
		PrintErr("No network interfaces specified.\n")
		fmt.Println(fmt.Sprintf("Usage example: %s -i eth0 -l 0.0.0.0:8080 -d ~/bwlog/\n", os.Args[0]))
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if config.Database == "" {
		PrintErr("No database directory specified.\n")
		fmt.Println(fmt.Sprintf("Usage example: %s -i eth0 -l 0.0.0.0:8080 -d ~/bwlog/\n", os.Args[0]))
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if bauth != "" {
		err := ReadBasicAuth(bauth)
		if err != nil {
			PrintErr(fmt.Sprintf("Cannot read authentication file %s.\n", bauth))
			os.Exit(1)
		}

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

		// stats controller
		http.HandleFunc("/stats/", BasicAuth(func(w http.ResponseWriter, r *http.Request) {
			statsController(w, r, config)
		}))

		// websocket route
		http.HandleFunc("/stream", BasicAuth(func(w http.ResponseWriter, r *http.Request) {
			streamController(w, r, config)
		}))

		http.HandleFunc("/", BasicAuth(func(w http.ResponseWriter, r *http.Request) {
			gziphandler.GzipHandler(http.FileServer(statikFS)).ServeHTTP(w, r)
		}))

		PrintInfo(fmt.Sprintf("HTTP listening on %s", config.Listen))

		log.Fatal(http.ListenAndServe(config.Listen, nil))
	}()

	// Stats daemon
	BWLogger(config)
}

// BasicAuth uses MySQL login details
func BasicAuth(handler http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if bauser != "" && bapass != "" {
			user, pass, ok := r.BasicAuth()

			if !ok {
				BasicAuthResponse(w)
				return
			}

			if user != bauser || pass != bapass {
				BasicAuthResponse(w)
				return
			}
		}

		// pass on request
		handler(w, r)
	}
}

// BasicAuthResponse returns an basic auth response to the browser
func BasicAuthResponse(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Login"`)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Unauthorised.\n"))
}

// ReadBasicAuth opens a file and sets a global user/password if a line with two words is found
func ReadBasicAuth(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := scanner.Text()
		if l != "" {
			words := strings.Fields(l)
			if len(words) == 2 {
				bauser = words[0]
				bapass = words[1]
			}
		}
	}
	err = scanner.Err()
	return err
}
