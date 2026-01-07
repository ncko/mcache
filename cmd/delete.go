package cmd

import (
	"fmt"
	"os"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <key>",
	Short: "Delete a key from memcached",
	Long: `Delete a key from the memcached server.

Example:
  mcache delete mykey
  mcache delete mykey --server=192.168.1.100 --port=11211`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		// Create memcached client
		serverAddr := fmt.Sprintf("%s:%d", server, port)
		mc := memcache.New(serverAddr)

		// Delete the key
		if err := mc.Delete(key); err != nil {
			if err == memcache.ErrCacheMiss {
				fmt.Fprintf(os.Stderr, "Error: key '%s' not found\n", key)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Error deleting key '%s' from %s: %v\n", key, serverAddr, err)
			os.Exit(1)
		}

		fmt.Printf("Deleted key '%s'\n", key)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
