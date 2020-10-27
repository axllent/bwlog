package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gobuffalo/packr"
	"github.com/gorilla/websocket"
)

// JSONReturn struct
type JSONReturn struct {
	If string
	Rx int64
	Tx int64
}

// Statistic struct
type Statistic struct {
	Date string
	RX   int64
	TX   int64
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// StartHTTP starts and HTTP server
func StartHTTP() {
	go func() {
		box := packr.NewBox("../web")

		// stats controller
		http.HandleFunc("/stats/", basicAuthWrapper(func(w http.ResponseWriter, r *http.Request) {
			statsController(w, r)
		}))

		// websocket route
		http.HandleFunc("/stream", basicAuthWrapper(func(w http.ResponseWriter, r *http.Request) {
			streamController(w, r)
		}))

		// everything else handled by static files
		http.HandleFunc("/", basicAuthWrapper(func(w http.ResponseWriter, r *http.Request) {
			gziphandler.GzipHandler(http.FileServer(box)).ServeHTTP(w, r)
		}))

		if Config.SSLCert != "" && Config.SSLKey != "" {
			fmt.Println("HTTPS listening on", Config.Listen)
			log.Fatal(http.ListenAndServeTLS(Config.Listen, Config.SSLCert, Config.SSLKey, nil))
		} else {
			fmt.Println("HTTP listening on", Config.Listen)
			log.Fatal(http.ListenAndServe(Config.Listen, nil))
		}
	}()
}

func streamController(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Print(err)
		}
		return
	}
	wsReader(conn)
}

// wsReader websocket reader for live runtime stats
func wsReader(ws *websocket.Conn) {
	ws.SetReadLimit(512)

	ticker := time.NewTicker(time.Second)

	// get initial readings
	last := make(map[string]InterfaceLog, len(Config.Interfaces))
	for _, nwIf := range Config.Interfaces {
		rx, tx, _ := IFStat(nwIf)
		last[nwIf] = InterfaceLog{RX: rx, TX: tx}
	}

	for range ticker.C {
		output := make([]JSONReturn, len(Config.Interfaces))

		for i, nwIf := range Config.Interfaces {
			prev, _ := last[nwIf]
			rx, tx, _ := IFStat(nwIf)
			if rx < prev.RX || tx < prev.TX {
				last[nwIf] = InterfaceLog{RX: rx, TX: tx}
				output[i] = JSONReturn{nwIf, 0, 0}
				continue
			}
			output[i] = JSONReturn{nwIf, (rx - prev.RX) / 1024, (tx - prev.TX) / 1024}
			last[nwIf] = InterfaceLog{RX: rx, TX: tx}
		}

		b, _ := json.Marshal(output)
		if err := ws.WriteMessage(websocket.TextMessage, b); err != nil {
			return
		}
	}
}

// Handles /stats/<nwif>(/<date>)
func statsController(w http.ResponseWriter, r *http.Request) {
	// update stats so they are current
	SyncNwInterfaces()
	re, _ := regexp.Compile(`/stats/([a-z0-9\-]+)/?(\d\d\d\d\-\d\d)?`)
	matches := re.FindStringSubmatch(r.URL.String())

	if len(matches) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - invalid URL!"))
		return
	}

	d, found := DB[matches[1]]
	if !(found) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - unknown interface!"))
		return
	}

	statsMonth := string(matches[2])

	data := []Stat{}
	if statsMonth != "" {
		data = []Stat{}
		for _, daily := range d.Daily {
			if strings.HasPrefix(daily.Date, statsMonth) {
				data = append(data, daily)
			}
		}
		data = reverseStats(data)
	} else {
		data = reverseStats(d.Monthly)
	}

	results, err := json.Marshal(data)
	if err != nil {
		log.Fatal("Cannot encode to JSON ", err)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", string(results))
}

// Reversestats returns a []Stat in reversed order
func reverseStats(input []Stat) []Stat {
	if len(input) == 0 {
		return input
	}
	return append(reverseStats(input[1:]), input[0])
}
