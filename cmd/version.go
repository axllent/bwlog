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

var (
	// Version is the default application version, updated on release
	Version = "dev"

	// Repo on Github for updater
	Repo = "axllent/bwlog"

	// RepoBinaryName on Github for updater
	RepoBinaryName = "bwlog"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the app version & update information",
	Long:  `Display the app version & update information.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(fmt.Sprintf("Version: %s", Version))
		latest, _, _, err := ghru.Latest(Repo, RepoBinaryName)
		if err == nil && ghru.GreaterThan(latest, Version) {
			fmt.Printf("Update available: %s\nRun `%s update` to update.\n", latest, os.Args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
