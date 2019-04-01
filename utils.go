package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

// Return array of rx, tx from network interface
func readStats(nwIf string) (int64, int64, error) {
	rx := fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", nwIf)
	tx := fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", nwIf)

	idata, err := ioutil.ReadFile(rx)
	if err != nil {
		return 0, 0, err
	}
	istr := strings.Trim(string(idata), "\n") // trim string
	received, err := strconv.ParseInt(istr, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	tdata, err := ioutil.ReadFile(tx)
	if err != nil {
		return 0, 0, err
	}
	tstr := strings.Trim(string(tdata), "\n") // trim string
	sent, err := strconv.ParseInt(tstr, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	// fmt.Println(received, sent)

	return received, sent, nil
}
