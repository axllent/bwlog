package main

import (
	"fmt"
	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"log"
	"time"
)

func LogStats(config Config) {

	conn, err := sqlite3.Open(config.Database)
	if err != nil {
		log.Fatal(err)
	}

	// prepare / create the table if necessary
	stmt, err := conn.Prepare(`SELECT name FROM sqlite_master WHERE type='table' AND name='?'`, "Daily")
	if err != nil {
		conn.Exec(`CREATE TABLE [Daily] ([Day] DATE, [Interface] VARCHAR (10), [RX] INTEGER, [TX] INTEGER)`)
		conn.Exec(`CREATE TABLE [Daily] ([Day] DATE, [Interface] VARCHAR (10), [RX] INTEGER, [TX] INTEGER)`)
		conn.Exec(`CREATE TABLE [Monthly] ([Month] DATE, [Interface] VARCHAR (10), [RX] INTEGER, [TX] INTEGER)`)
		conn.Exec(`CREATE UNIQUE INDEX [idx_Daily] ON [Daily] ([Day], [Interface])`)
		conn.Exec(`CREATE UNIQUE INDEX [idx_Monthly] ON [Monthly] ([Month], [Interface])`)
	}
	defer stmt.Close()

	fmt.Println(sqlDate())

	conn.Close()

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

	// loop the functionality
	ticker := time.NewTicker(1000 * time.Millisecond)
	for range ticker.C {
		conn, _ := sqlite3.Open(config.Database)

		for i := 0; i < len(config.Interfaces); i++ {
			if rx, tx, err := readStats(config.Interfaces[i]); err == nil {
				rec := (rx - stats[i][0]) / 1024
				snd := (tx - stats[i][1]) / 1024
				stats[i][0] = rx
				stats[i][1] = tx

				today := sqlDate()

				// fmt.Println(fmt.Sprintf(`INSERT OR IGNORE INTO Daily VALUES(%s, %s, 0, 0); UPDATE Daily SET RX=RX+%d, TX=TX+%d WHERE Day=%s AND Interface=%s`, today, config.Interfaces[i], rec, snd, today, config.Interfaces[i]))

				// err = conn.Exec(`INSERT OR IGNORE INTO Daily VALUES(?, ?, 0, 0); UPDATE Daily SET RX=RX+?, TX=TX+? WHERE Day=? AND Interface=?`, today, config.Interfaces[i], rec, snd, today, config.Interfaces[i])
				// err = conn.Exec(`INSERT OR IGNORE INTO Daily VALUES(?, ?, 0, 0);`)// UPDATE Daily SET RX=RX+?, TX=TX+? WHERE Day=? AND Interface=?`, "2019-0-01", "eth0", "100", "100", "2019-0-01", "eth0")
				conn.Exec(`INSERT OR IGNORE INTO Daily VALUES(?, ?, 0, 0)`, today, config.Interfaces[i])
				err = conn.Exec(`UPDATE Daily SET RX=RX+?, TX=TX+? WHERE Day=? AND Interface=?`, rec, snd, today, config.Interfaces[i])
				if err != nil {
					log.Fatal(err)
				}
				// defer stmt.Close()
				fmt.Println(config.Interfaces[i], rec, snd)
			}
		}

		conn.Close()
	}
}
