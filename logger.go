/**
 * Logger
 */
package main

import (
	"fmt"
	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"log"
	"time"
)

func bwLogger(config Config) {

	conn, err := sqlite3.Open(config.Database)
	if err != nil {
		log.Fatal(err)
	}

	// create the tables if necessary (silently fails if they exist)
	conn.Exec(`CREATE TABLE [Daily] ([Day] DATE, [Interface] VARCHAR (10), [RX] INTEGER, [TX] INTEGER)`)
	conn.Exec(`CREATE TABLE [Daily] ([Day] DATE, [Interface] VARCHAR (10), [RX] INTEGER, [TX] INTEGER)`)
	conn.Exec(`CREATE TABLE [Monthly] ([Month] DATE, [Interface] VARCHAR (10), [RX] INTEGER, [TX] INTEGER)`)
	conn.Exec(`CREATE UNIQUE INDEX [idx_Daily] ON [Daily] ([Day], [Interface])`)
	conn.Exec(`CREATE UNIQUE INDEX [idx_Monthly] ON [Monthly] ([Month], [Interface])`)
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
	ticker := time.NewTicker(time.Duration(config.Save*1000) * time.Millisecond)

	for range ticker.C {
		conn, _ := sqlite3.Open(config.Database)

		currentTime := time.Now()
		sqlDay := currentTime.Format("2006-01-02")
		sqlMonth := currentTime.Format("2006-01")

		for i := 0; i < len(config.Interfaces); i++ {
			if rx, tx, err := readStats(config.Interfaces[i]); err == nil {
				in := (rx - stats[i][0]) / 1024
				out := (tx - stats[i][1]) / 1024

				if in > 0 && out > 0 {
					// fmt.Println("+ Logging", config.Interfaces[i], in, out)
					// Daily totals
					err = conn.Exec(`INSERT OR IGNORE INTO Daily
						VALUES(?, ?, 0, 0)	ON CONFLICT(Day, Interface)
						DO UPDATE SET RX=RX+?, TX=TX+?`,
						sqlDay, config.Interfaces[i], in, out)
					if err != nil {
						fmt.Println(err)
						continue
					}
					// Monthly totals
					conn.Exec(`INSERT OR IGNORE INTO Monthly VALUES(?, ?, 0, 0)
						ON CONFLICT(Month, Interface) DO UPDATE SET RX=RX+?, TX=TX+?`,
						sqlMonth, config.Interfaces[i], in, out)
				}

				// set new totals
				stats[i][0] = rx
				stats[i][1] = tx
			}
		}

		conn.Close()
	}
}
