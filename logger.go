package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// BWLogger periodically saves the current stats to CSV
func BWLogger(config Config) {

	PrintInfo(fmt.Sprintf("BWLog: Logging %s to %s", strings.Join(config.Interfaces, ","), config.Database))

	// create stats array
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

	for ; true; <-ticker.C {
		currentTime := time.Now()
		csvDay := currentTime.Format("2006-01-02")
		csvMonth := currentTime.Format("2006-01")

		for i := 0; i < len(config.Interfaces); i++ {
			if rx, tx, err := readStats(config.Interfaces[i]); err == nil {
				in := (rx - stats[i][0]) / 1024
				out := (tx - stats[i][1]) / 1024

				dailyname := fmt.Sprintf("%s_daily.csv", config.Interfaces[i])
				dailydb := filepath.Join(config.Database, dailyname)

				CreateDB(dailydb, "Day")

				monthlyname := fmt.Sprintf("%s_monthly.csv", config.Interfaces[i])
				monthlydb := filepath.Join(config.Database, monthlyname)

				CreateDB(monthlydb, "Month")

				err = LogToDB(dailydb, csvDay, in, out)
				if err != nil {
					fmt.Println(err)
					continue
				}

				err = LogToDB(monthlydb, csvMonth, in, out)
				if err != nil {
					fmt.Println(err)
					continue
				}

				// set new totals
				stats[i][0] = rx
				stats[i][1] = tx
			}
		}
	}
}

// CreateDB creates a new CSV file and append headers
func CreateDB(datafile string, datehdr string) {
	_, err := os.Stat(datafile)
	if err != nil {
		// file does not exist, create
		// fmt.Println(fmt.Sprintf("Creating new database file for %s", datafile))
		f, err := os.OpenFile(datafile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
			return
		}
		w := csv.NewWriter(f)
		w.Write([]string{datehdr, "RX", "TX"})
		w.Flush()
	}
}

// LogToDB will open, read and append to log file
func LogToDB(path string, date string, rx int64, tx int64) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return err
	}

	f.Close()

	match := false

	// read bottom to top
	for i := len(rows) - 1; i > 0; i-- {
		if rows[i][0] == date {
			match = true
			newrx := AddInt64ToString(rows[i][1], rx)
			rows[i][1] = newrx
			newtx := AddInt64ToString(rows[i][2], tx)
			rows[i][2] = newtx
			continue
		}
	}

	if !match {
		rows = append(rows, []string{date, strconv.FormatInt(rx, 10), strconv.FormatInt(tx, 10)})
	}

	w, err := os.Create(path)
	if err != nil {
		return err
	}

	err = csv.NewWriter(w).WriteAll(rows)
	w.Close()
	if err != nil {
		return err
	}

	return nil
}

// AddInt64ToString adds an int64 to a string, and return as as string for csv
func AddInt64ToString(str string, val int64) string {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return strconv.FormatInt(val, 10)
	}

	i = i + val
	return strconv.FormatInt(i, 10)
}
