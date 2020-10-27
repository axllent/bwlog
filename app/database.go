package app

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"sync"
)

var (
	// DB is the global database variable
	DB = make(map[string]DBStruct)

	db     sync.Mutex
	write  sync.Mutex
	ifsync sync.Mutex

	dayRegex   = regexp.MustCompile(`^\d\d\d\d\-\d\d\-\d\d$`)
	monthRegex = regexp.MustCompile(`^\d\d\d\d\-\d\d$`)
)

// Stat struct stores each statistic
type Stat struct {
	Date string
	RX   int64
	TX   int64
}

// DBStruct represents the database layout for an interface
type DBStruct struct {
	Daily   []Stat
	Monthly []Stat
}

// AddStat will set a statistic to the database.
func AddStat(nwIf, date string, rx, tx int64) {
	st, err := statType(date)
	if err != nil {
		// fmt.Println("Error:", err)
		return
	}
	db.Lock()

	d, ok := DB[nwIf]
	if !ok {
		d = DBStruct{
			Daily:   []Stat{},
			Monthly: []Stat{},
		}
	}

	if st == "Day" {
		found := false
		for i, v := range DB[nwIf].Daily {
			if v.Date == date {
				found = true
				DB[nwIf].Daily[i].RX = v.RX + rx
				DB[nwIf].Daily[i].TX = v.TX + tx
			}
		}

		if !found {
			d.Daily = append(DB[nwIf].Daily, Stat{Date: date, RX: rx, TX: tx})
		}
	} else if st == "Month" {
		found := false
		for i, v := range DB[nwIf].Monthly {
			if v.Date == date {
				found = true
				DB[nwIf].Monthly[i].RX = v.RX + rx
				DB[nwIf].Monthly[i].TX = v.TX + tx
			}
		}

		if !found {
			d.Monthly = append(DB[nwIf].Monthly, Stat{Date: date, RX: rx, TX: tx})
		}
	}
	DB[nwIf] = d

	db.Unlock()
}

// StatType returns whether a date is a month (yyyy-mm) or a day (yyyy-mm-dd)
func statType(date string) (string, error) {
	if dayRegex.MatchString(date) {
		return "Day", nil
	}
	if monthRegex.MatchString(date) {
		return "Month", nil
	}

	return "", fmt.Errorf("%s is an invalid date format", date)
}

// LoadStats will load any existing CSV files for a matching network interface
func LoadStats(nwIf string) error {
	for _, logType := range []string{"daily", "monthly"} {

		filename := fmt.Sprintf("%s_%s.csv", nwIf, logType)
		file := path.Join(Config.DatabaseDir, filename)

		f, err := os.Open(file)
		if err != nil {
			return err
		}

		rows, err := csv.NewReader(f).ReadAll()
		if err != nil {
			return err
		}

		f.Close()

		for _, row := range rows {
			rx, _ := strconv.ParseInt(row[1], 10, 64)
			tx, _ := strconv.ParseInt(row[2], 10, 64)
			AddStat(nwIf, row[0], rx, tx)
		}
	}

	d, ok := DB[nwIf]
	if !ok {
		return nil
	}

	sort.SliceStable(d.Daily, func(i, j int) bool {
		return d.Daily[i].Date < d.Daily[j].Date
	})

	sort.SliceStable(d.Monthly, func(i, j int) bool {
		return d.Monthly[i].Date < d.Monthly[j].Date
	})

	return nil
}

// SaveStats will save the running stats to CSV files
func SaveStats() {
	write.Lock()
	for _, nwIf := range Config.Interfaces {
		dailyname := fmt.Sprintf("%s_daily.csv", nwIf)
		daily := filepath.Join(Config.DatabaseDir, dailyname)

		if err := writeCSV(daily, "Day", DB[nwIf].Daily); err != nil {
			fmt.Println(err)
		}

		monthlyname := fmt.Sprintf("%s_monthly.csv", nwIf)
		monthly := filepath.Join(Config.DatabaseDir, monthlyname)

		if err := writeCSV(monthly, "Month", DB[nwIf].Monthly); err != nil {
			fmt.Println(err)
		}
	}
	write.Unlock()
}

func writeCSV(file, title string, data []Stat) error {
	var csvData = [][]string{}
	csvData = append(csvData, []string{title, "RX", "TX"})
	for _, stat := range data {
		csvData = append(csvData, []string{
			stat.Date,
			strconv.FormatInt(stat.RX, 10),
			strconv.FormatInt(stat.TX, 10),
		})
	}

	w, err := os.Create(file)
	if err != nil {
		return err
	}
	defer w.Close()

	return csv.NewWriter(w).WriteAll(csvData)
}
