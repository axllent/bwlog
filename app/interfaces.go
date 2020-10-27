package app

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/axllent/bwlog/utils"
	"github.com/spf13/cobra"
)

// InterfaceLog struct stores each interface's previous stats to calculate difference
type InterfaceLog struct {
	RX int64
	TX int64
}

// InitInterfaces will set a slice of config interfaces from the cmd
func InitInterfaces(cmd *cobra.Command) error {
	interfaces, _ := cmd.Flags().GetString("interfaces")
	nwIf := strings.Split(interfaces, ",")
	for _, v := range nwIf {
		trimv := strings.TrimSpace(v)
		if trimv != "" {
			Config.Interfaces = append(Config.Interfaces, trimv)
		}
	}
	if len(Config.Interfaces) == 0 {
		return errors.New("No network interfaces set")
	}

	return nil
}

// IFStat returns the current of rx & tx values from network interface
func IFStat(nwIf string) (int64, int64, error) {
	rx := fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", nwIf)
	tx := fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", nwIf)

	// network doesn't exist
	if !utils.IsFile(rx) || !utils.IsFile(tx) {
		return 0, 0, nil
	}

	rxdata, err := ioutil.ReadFile(rx)
	if err != nil {
		return 0, 0, err
	}

	// convert string to int64
	received, err := strconv.ParseInt(strings.TrimSpace(string(rxdata)), 10, 64)
	if err != nil {
		return 0, 0, err
	}

	txdata, err := ioutil.ReadFile(tx)
	if err != nil {
		return 0, 0, err
	}

	// convert string to int64
	sent, err := strconv.ParseInt(strings.TrimSpace(string(txdata)), 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return received, sent, nil
}

// ifStatMap is the last recorded interface totals
var ifStatMap = make(map[string]InterfaceLog)

// SyncNwInterfaces will log each interface to the global DB if either the RX or TX > 0
func SyncNwInterfaces() {
	ifsync.Lock()
	for _, nwIf := range Config.Interfaces {
		curRX, curTX, err := IFStat(nwIf)
		if err != nil {
			fmt.Println(err)
			continue
		}

		prev, exists := ifStatMap[nwIf]
		if !exists {
			ifStatMap[nwIf] = InterfaceLog{
				RX: curRX,
				TX: curTX,
			}
			// set the values and don't log
			continue
		}

		if curRX < prev.RX || curTX < prev.TX {
			ifStatMap[nwIf] = InterfaceLog{
				RX: curRX,
				TX: curTX,
			}
			// reset the values and don't log
			continue
		}

		rx := (curRX - prev.RX) / 1024
		tx := (curTX - prev.TX) / 1024

		day := time.Now().Format("2006-01-02")
		month := time.Now().Format("2006-01")

		AddStat(nwIf, day, rx, tx)
		AddStat(nwIf, month, rx, tx)

		ifStatMap[nwIf] = InterfaceLog{
			RX: curRX,
			TX: curTX,
		}
	}

	ifsync.Unlock()
}
