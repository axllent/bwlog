package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/axllent/bwlog/app"
	"github.com/axllent/bwlog/utils"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "bwlog -i eth0 -d ~/bwlog/",
	Short:         "BWLog: A lightweight bandwidth logger",
	SilenceErrors: true, // suppress duplicate error on error
	Args:          cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initConfig(cmd); err != nil {
			return err
		}

		for _, nwIf := range app.Config.Interfaces {
			app.LoadStats(nwIf)
		}

		app.SyncNwInterfaces()

		sigs := make(chan os.Signal, 1)
		// catch all signals since not explicitly listing
		// Program that will listen to the SIGINT and SIGTERM
		// SIGINT will listen to CTRL-C.
		// SIGTERM will be caught if kill command executed
		signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
		// method invoked upon seeing signal
		go func() {
			s := <-sigs
			fmt.Printf("Got %s signal, saving data & shutting down...\n", s)
			app.SyncNwInterfaces()
			app.SaveStats()
			os.Exit(0)
		}()

		app.StartHTTP()

		ticker := time.NewTicker(app.Config.SaveInterval)

		for range ticker.C {
			app.SyncNwInterfaces()
			app.SaveStats()
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&app.Config.DatabaseDir, "database", "d", "./", "Database directory to save CSV files")
	rootCmd.Flags().StringVarP(&app.Config.Listen, "listen", "l", "0.0.0.0:8080", "Interface & port to listen on")
	rootCmd.Flags().StringVar(&app.Config.SSLCert, "sslcert", "", "SSL certificate (must be used together with --sslkey)")
	rootCmd.Flags().StringVar(&app.Config.SSLKey, "sslkey", "", "SSL key (must be used together with --sslcert)")
	rootCmd.Flags().StringP("interfaces", "i", "", "Interfaces to monitor, comma separated eg: eth0,eth1")
	rootCmd.Flags().StringP("password", "p", "", "Auth password file (must contain a single \"<user> <pass>\")")
	rootCmd.Flags().StringP("save", "s", "60s", "How often to save the database to disk. Examples: 30s, 5m, 1h")

	rootCmd.MarkFlagRequired("interfaces")

	rootCmd.Flags().BoolP("help", "h", false, "override help so we can hide it")
	rootCmd.Flags().MarkHidden("help")
}

// initConfig reads in config file and ENV variables if set.
func initConfig(cmd *cobra.Command) error {
	if err := app.InitInterfaces(cmd); err != nil {
		return err
	}

	passwordFile, _ := cmd.Flags().GetString("password")
	if passwordFile != "" {
		if err := app.BasicAuthFromFile(passwordFile); err != nil {
			return err
		}
	}

	save, _ := cmd.Flags().GetString("save")
	intrvl, err := time.ParseDuration(save)
	if err != nil {
		return err
	}
	app.Config.SaveInterval = intrvl

	if !utils.IsDir(app.Config.DatabaseDir) {
		if err := os.Mkdir(app.Config.DatabaseDir, 0755); err != nil {
			return err
		}
	}

	return nil
}
