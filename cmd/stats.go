package cmd

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show memcached server statistics",
	Long: `Display statistics from the memcached server.

Example:
  mcache stats
  mcache stats --server=192.168.1.100 --port=11211`,
	Run: func(cmd *cobra.Command, args []string) {
		serverAddr := fmt.Sprintf("%s:%d", server, port)

		// Connect to memcached
		conn, err := net.Dial("tcp", serverAddr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to %s: %v\n", serverAddr, err)
			os.Exit(1)
		}
		defer conn.Close()

		// Send stats command
		fmt.Fprintf(conn, "stats\r\n")

		// Read response
		stats := make(map[string]string)
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "END" {
				break
			}
			if strings.HasPrefix(line, "STAT ") {
				parts := strings.SplitN(line[5:], " ", 2)
				if len(parts) == 2 {
					stats[parts[0]] = parts[1]
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stats: %v\n", err)
			os.Exit(1)
		}

		// Display stats
		fmt.Printf("Server: %s\n\n", serverAddr)

		// Sort keys for consistent output
		keys := make([]string, 0, len(stats))
		for k := range stats {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			fmt.Printf("  %s: %s\n", k, stats[k])
		}
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
