package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

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

func streamController(w http.ResponseWriter, r *http.Request, config Config) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Print(err)
		}
		return
	}
	wsReader(conn, config)
}

// wsReader sebsocket reader
func wsReader(ws *websocket.Conn, config Config) {
	ws.SetReadLimit(512)

	// create stats
	stats := make([][]int64, len(config.Interfaces))

	for i := 0; i < len(config.Interfaces); i++ {

		if rx, tx, err := readStats(config.Interfaces[i]); err == nil {
			// create stats for each interface
			stats[i] = make([]int64, 2)
			stats[i][0] = rx
			stats[i][1] = tx
		}
	}

	ticker := time.NewTicker(1000 * time.Millisecond)

	defer func() {
		// log.Print("Websocket disconnected")
		ticker.Stop()
		ws.Close()
	}()

	for range ticker.C {
		output := make([]JSONReturn, len(config.Interfaces))
		for i := 0; i < len(config.Interfaces); i++ {
			if rx, tx, err := readStats(config.Interfaces[i]); err == nil {
				in := (rx - stats[i][0]) / 1024
				out := (tx - stats[i][1]) / 1024
				stats[i][0] = rx
				stats[i][1] = tx
				m := JSONReturn{config.Interfaces[i], in, out}
				output[i] = m
			} else {
				m := JSONReturn{config.Interfaces[i], 0, 0}
				output[i] = m
			}
		}

		b, _ := json.Marshal(output)
		if err := ws.WriteMessage(websocket.TextMessage, b); err != nil {
			return
		}
	}
}

// Handles /stats/<nwif>(/<date>)?
func statsController(w http.ResponseWriter, r *http.Request, config Config) {
	re, _ := regexp.Compile(`/stats/([a-z0-9\-]+)/?(\d\d\d\d\-\d\d)?`)
	matches := re.FindStringSubmatch(r.URL.String())

	if len(matches) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - invalid URL!"))
		return
	}

	nwif := string(matches[1])

	if !InArray(nwif, config.Interfaces) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - unknown interface!"))
	}

	statsMonth := string(matches[2])

	var filename = string("")

	if statsMonth != "" {
		// daily statictics
		filename = fmt.Sprintf("%s_daily.csv", nwif)
	} else {
		// monthly statictics
		filename = fmt.Sprintf("%s_monthly.csv", nwif)
	}

	datafile := filepath.Join(config.Database, filename)

	stats, _ := ReturnRetults(datafile, statsMonth)

	results, err := json.Marshal(stats)
	if err != nil {
		log.Fatal("Cannot encode to JSON ", err)
	}

	// fmt.Println(string(results))
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", string(results))
}

// ReturnRetults opens a data file, and returns a slice of results in reverse order
func ReturnRetults(datafile string, date string) ([]Statistic, error) {
	var stats []Statistic

	f, err := os.Open(datafile)
	if err != nil {
		return stats, err
	}

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return stats, err
	}

	f.Close()

	// read bottom to top
	for i := len(rows) - 1; i > 0; i-- {
		if date == "" || strings.Contains(rows[i][0], date) {
			rx, _ := strconv.ParseInt(rows[i][1], 10, 64)
			tx, _ := strconv.ParseInt(rows[i][2], 10, 64)

			stats = append(stats, Statistic{
				Date: rows[i][0],
				RX:   rx,
				TX:   tx,
			})
		}
	}

	return stats, nil
}

// InArray is a php-like InArray() function
func InArray(x string, a []string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
