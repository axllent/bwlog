package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

// Return int64 array of rx & tx values from network interface
func readStats(nwIf string) (int64, int64, error) {
	rx := fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", nwIf)
	tx := fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", nwIf)

	idata, err := ioutil.ReadFile(rx)
	if err != nil {
		return 0, 0, err
	}
	// trim string
	istr := strings.Trim(string(idata), "\n")
	// convert string to int64
	received, err := strconv.ParseInt(istr, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	tdata, err := ioutil.ReadFile(tx)
	if err != nil {
		return 0, 0, err
	}
	// trim string
	tstr := strings.TrimSpace(string(tdata))
	// convert string to int64
	sent, err := strconv.ParseInt(tstr, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return received, sent, nil
}

// PrintInfo prints info message
func PrintInfo(str string) {
	fmt.Println(str)
}

// PrintErr prints error in red
func PrintErr(str string) {
	fmt.Println(fmt.Sprintf("\033[1;31m%s\033[0m", str))
}
