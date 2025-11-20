package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	server string
	port   int
)

var rootCmd = &cobra.Command{
	Use:   "mcache",
	Short: "A CLI tool for interacting with remote memcached servers",
	Long: `mcache is a command-line interface tool that allows you to
interact with remote memcached servers. You can retrieve, set,
and manage cached data directly from your terminal.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags for all commands
	rootCmd.PersistentFlags().StringVarP(&server, "server", "s", "localhost", "Memcached server address")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 11211, "Memcached server port")
}
