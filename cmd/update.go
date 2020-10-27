/*
Copyright © 2020-Now() Ralph Slooten
This file is part of a CLI application.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/axllent/ghru"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update bwlog to the latest version",
	Long:  `Update bwlog to the latest version.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return updateApp()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func updateApp() error {
	rel, err := ghru.Update(Repo, RepoBinaryName, Version)
	if err != nil {
		return err
	}
	fmt.Printf("Updated %s to version %s\n", os.Args[0], rel)
	return nil
}
