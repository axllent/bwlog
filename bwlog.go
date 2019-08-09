package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/NYTimes/gziphandler"
	"github.com/axllent/gitrel"
	"github.com/gobuffalo/packr"
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

	var bauth, interfaces, sslcert, sslkey string
	var update, showversion bool
	flag.StringVar(&bauth, "p", "", "basic auth password file (must contain a single <user>:<pass>)")
	flag.StringVar(&interfaces, "i", "", "interfaces to monitor, comma separated eg: eth0,eth1")
	flag.StringVar(&config.Listen, "l", "0.0.0.0:8080", "port to listen on")
	flag.StringVar(&config.Database, "d", "", "database directory path")
	flag.StringVar(&sslcert, "sslcert", "", "Path to SSL certificate (must be used together with sslkey)")
	flag.StringVar(&sslkey, "sslkey", "", "Path to private SSL key (must be used together with sslcert)")
	flag.IntVar(&config.Save, "s", 60, "save to database every X seconds")
	flag.BoolVar(&update, "u", false, "update to latest release")
	flag.BoolVar(&showversion, "v", false, "show version number")

	flag.Usage = func() {
		fmt.Printf("BWLog %s: A lightweight bandwidth logger.\n\n", version)
		fmt.Printf("Usage example: %s -i eth0 -l 0.0.0.0:8080 -d ~/bwlog/\n\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
	}

	flag.Usage = func() {
		fmt.Printf("BWLog %s: A lightweight bandwidth logger.\n\n", version)
		fmt.Printf("Usage example: %s -i eth0 -l 0.0.0.0:8080 -d ~/bwlog.sqlite\n\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	config.Interfaces = strings.Split(interfaces, ",")

	if showversion {
		fmt.Printf("Version: %s\n", version)
		latest, _, _, err := gitrel.Latest("axllent/bwlog", "bwlog")
		if err == nil && latest != version {
			fmt.Printf("Update available: %s\nRun `%s -u` to update.\n", latest, os.Args[0])
		}
		return
	}

	if update {
		rel, err := gitrel.Update("axllent/bwlog", "bwlog", version)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Updated %s to version %s", os.Args[0], rel)
		return
	}

	if interfaces == "" {
		PrintErr("No network interfaces specified.\n")
		fmt.Printf("Usage example: %s -i eth0 -l 0.0.0.0:8080 -d ~/bwlog/\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if config.Database == "" {
		PrintErr("No database directory specified.\n")
		fmt.Printf("Usage example: %s -i eth0 -l 0.0.0.0:8080 -d ~/bwlog/\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if bauth != "" {
		err := ReadBasicAuth(bauth)
		if err != nil {
			fmt.Printf("Cannot read authentication file %s.\n", bauth)
			os.Exit(1)
		}

	}

	dbinfo, err := os.Stat(config.Database)
	if err != nil {
		fmt.Printf("%s does not exist, exiting\n", config.Database)
		os.Exit(1)
	}

	if !dbinfo.IsDir() {
		fmt.Printf("%s is not a directory, exiting\n", config.Database)
		os.Exit(1)
	}

	sigs := make(chan os.Signal, 1)
	// catch all signals since not explicitly listing
	// Program that will listen to the SIGINT and SIGTERM
	// SIGINT will listen to CTRL-C.
	// SIGTERM will be caught if kill command executed
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	// method invoked upon seeing signal
	go func() {
		s := <-sigs
		fmt.Printf("Got %s signal, saving data & shutting down...\n", s)
		config.SaveStats()
		os.Exit(1)
	}()

	// Start new thread for httpd
	go func() {
		box := packr.NewBox("./web")

		// stats controller
		http.HandleFunc("/stats/", BasicAuth(func(w http.ResponseWriter, r *http.Request) {
			statsController(w, r, config)
		}))

		// websocket route
		http.HandleFunc("/stream", BasicAuth(func(w http.ResponseWriter, r *http.Request) {
			streamController(w, r, config)
		}))

		// everything else handled by static files
		http.HandleFunc("/", BasicAuth(func(w http.ResponseWriter, r *http.Request) {
			gziphandler.GzipHandler(http.FileServer(box)).ServeHTTP(w, r)
		}))

		if sslcert != "" && sslkey != "" {
			PrintInfo(fmt.Sprintf("HTTPS listening on %s", config.Listen))
			log.Fatal(http.ListenAndServeTLS(config.Listen, sslcert, sslkey, nil))
		} else {
			PrintInfo(fmt.Sprintf("HTTP listening on %s", config.Listen))
			log.Fatal(http.ListenAndServe(config.Listen, nil))
		}
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
