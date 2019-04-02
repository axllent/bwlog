package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

type JsonReturn struct {
	If string
	Rx int64
	Tx int64
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
	log.Print("Websocket connected")
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
		log.Print("Websocket disconnected")
		ticker.Stop()
		ws.Close()
	}()

	for range ticker.C {
		// response := make([][]JsonReturn, len(config.Interfaces))
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
