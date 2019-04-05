package main

import (
	"encoding/json"
	"fmt"
	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"regexp"
	"time"
)

type JsonReturn struct {
	If string
	Rx int64
	Tx int64
}

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
	// log.Print("Websocket connected")
	wsReader(conn, config)
}

func wsReader(ws *websocket.Conn, config Config) {
	// defer ws.Close()
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
		output := make([]JsonReturn, len(config.Interfaces))
		for i := 0; i < len(config.Interfaces); i++ {
			if rx, tx, err := readStats(config.Interfaces[i]); err == nil {
				in := (rx - stats[i][0]) / 1024
				out := (tx - stats[i][1]) / 1024
				stats[i][0] = rx
				stats[i][1] = tx
				m := JsonReturn{config.Interfaces[i], in, out}
				output[i] = m
			} else {
				m := JsonReturn{config.Interfaces[i], 0, 0}
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

	if !in_array(nwif, config.Interfaces) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - unknown interface!"))
	}

	stats_month := string(matches[2])

	conn, _ := sqlite3.Open(config.Database)

	defer conn.Close()

	var stats []Statistic
	var stmt *sqlite3.Stmt

	// fmt.Println(matches)

	if stats_month != "" {
		// daily stats
		month := string(matches[2]) + "%"
		stmt, _ = conn.Prepare(`SELECT Day, RX, TX FROM Daily WHERE Interface = ? AND Day LIKE ? ORDER BY Day DESC`, nwif, month)
	} else {
		// monthlt stats
		stmt, _ = conn.Prepare(`SELECT Month, RX, TX FROM Monthly WHERE Interface = ? ORDER BY Month DESC`, nwif)
	}

	for {
		hasRow, _ := stmt.Step()
		if !hasRow {
			// The query is finished
			break
		}

		var month string
		var rx int64
		var tx int64
		stmt.Scan(&month, &rx, &tx)

		stats = append(stats, Statistic{
			Date: month,
			RX:   rx,
			TX:   tx,
		})
	}

	results, err := json.Marshal(stats)
	if err != nil {
		log.Fatal("Cannot encode to JSON ", err)
	}

	// fmt.Println(string(results))
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", string(results))
}

// php-like in_array() function
func in_array(x string, a []string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
