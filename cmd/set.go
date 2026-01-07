package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/spf13/cobra"
)

var (
	expiration int
	flags      uint32
	inputFile  string
)

var setCmd = &cobra.Command{
	Use:   "set <key> [value]",
	Short: "Set a key-value pair in memcached",
	Long: `Set a key-value pair in the memcached server.

The value can be provided as an argument, from stdin, or from a file.

Example:
  mcache set mykey "my value"
  mcache set mykey --input-file=data.txt
  echo "my value" | mcache set mykey
  mcache set mykey "cached data" --expiration=3600`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		var value []byte
		var err error

		// Determine value source: argument, file, or stdin
		if len(args) == 2 {
			value = []byte(args[1])
		} else if inputFile != "" {
			value, err = os.ReadFile(inputFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file '%s': %v\n", inputFile, err)
				os.Exit(1)
			}
		} else {
			// Check if stdin has data
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				value, err = io.ReadAll(os.Stdin)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
					os.Exit(1)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Error: no value provided. Use argument, --input-file, or pipe from stdin\n")
				os.Exit(1)
			}
		}

		// Create memcached client
		serverAddr := fmt.Sprintf("%s:%d", server, port)
		mc := memcache.New(serverAddr)

		// Set the item
		item := &memcache.Item{
			Key:        key,
			Value:      value,
			Flags:      flags,
			Expiration: int32(expiration),
		}

		if err := mc.Set(item); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting key '%s' on %s: %v\n", key, serverAddr, err)
			os.Exit(1)
		}

		fmt.Printf("Set key '%s' (%d bytes)\n", key, len(value))
	},
}

func init() {
	rootCmd.AddCommand(setCmd)

	setCmd.Flags().IntVarP(&expiration, "expiration", "e", 0, "Expiration time in seconds (0 = never expires)")
	setCmd.Flags().Uint32VarP(&flags, "flags", "F", 0, "Flags to store with the item")
	setCmd.Flags().StringVarP(&inputFile, "input-file", "i", "", "Read value from file")
}
