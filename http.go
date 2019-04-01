package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

type JsonReturn struct {
	If string
	Rx int64
	Tx int64
}

func rootController(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func streamController(w http.ResponseWriter, r *http.Request, config Config) {
	conn, err := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			// not a websocket request
			fmt.Println(err)
		}
		return
	}
	fmt.Println("+ Client connected")
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
		fmt.Println("- Client disconnected")
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
		fmt.Println("> Sending ws response")

		b, _ := json.Marshal(output)
		if err := ws.WriteMessage(websocket.TextMessage, b); err != nil {
			return
		}
	}
}

// func socketController(w http.ResponseWriter, r *http.Request, config Config) {
// 	conn, err := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
// 	if err != nil {
// 		if _, ok := err.(websocket.HandshakeError); !ok {
// 			// not a websocket request
// 			fmt.Println(err)
// 		}
// 		return
// 	}

// 	for {
// 		// Read message from browser
// 		_, msg, err := conn.ReadMessage()
// 		if err != nil {
// 			return
// 		}

// 		// Print the message to the console
// 		fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))
// 		fmt.Printf("sent: %s\n", config.Database)

// 		// Write message back to browser
// 		response := []byte(fmt.Sprintf("You wrote: %s", msg))

// 		if err = conn.WriteMessage(websocket.TextMessage, response); err != nil {
// 			return
// 		}
// 		// if err = conn.WriteMessage(msgType, msg); err != nil {
// 		// 	return
// 		// }
// 	}
// }
