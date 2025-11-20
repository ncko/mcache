package cmd

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/spf13/cobra"
)

var (
	outputFormat  string
	outputFile    string
	tryDecompress bool
	verbose       bool
	valueOnly     bool
)

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get the value of a key from memcached",
	Long: `Retrieve and display the contents of a specific key from the memcached server.

Example:
  mcache get mykey
  mcache get mykey --server=192.168.1.100 --port=11211`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		// Create memcached client
		serverAddr := fmt.Sprintf("%s:%d", server, port)
		mc := memcache.New(serverAddr)

		// Get the item from memcached
		item, err := mc.Get(key)
		if err != nil {
			if err == memcache.ErrCacheMiss {
				fmt.Fprintf(os.Stderr, "Error: key '%s' not found\n", key)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Error connecting to memcached at %s: %v\n", serverAddr, err)
			os.Exit(1)
		}

		// Process the value
		value := item.Value
		originalSize := len(value)
		wasDecompressed := false

		// Verbose: show first bytes of original data
		if verbose {
			fmt.Fprintf(os.Stderr, "[DEBUG] Original size: %d bytes\n", originalSize)
			fmt.Fprintf(os.Stderr, "[DEBUG] First 16 bytes (hex): %x\n", value[:min(16, len(value))])
			fmt.Fprintf(os.Stderr, "[DEBUG] Is gzipped: %v\n", isGzipped(value))
			fmt.Fprintf(os.Stderr, "[DEBUG] Is zlib: %v\n", isZlib(value))
			fmt.Fprintf(os.Stderr, "[DEBUG] Is valid UTF-8: %v\n", utf8.Valid(value))
		}

		// Try to decompress if requested or if data looks compressed
		if tryDecompress || isGzipped(value) || isZlib(value) {
			var decompressed []byte
			var err error

			// Try zlib first (more common in memcached)
			if isZlib(value) {
				decompressed, err = tryZlibDecompress(value)
				if verbose {
					if err == nil {
						fmt.Fprintf(os.Stderr, "[DEBUG] Zlib decompression successful!\n")
					} else {
						fmt.Fprintf(os.Stderr, "[DEBUG] Zlib decompression failed: %v\n", err)
					}
				}
			}

			// Try gzip if zlib failed or wasn't detected
			if err != nil || (!isZlib(value) && isGzipped(value)) {
				decompressed, err = tryGzipDecompress(value)
				if verbose {
					if err == nil {
						fmt.Fprintf(os.Stderr, "[DEBUG] Gzip decompression successful!\n")
					} else {
						fmt.Fprintf(os.Stderr, "[DEBUG] Gzip decompression failed: %v\n", err)
					}
				}
			}

			if err == nil && decompressed != nil {
				if verbose {
					fmt.Fprintf(os.Stderr, "[DEBUG] Decompressed size: %d bytes\n", len(decompressed))
					fmt.Fprintf(os.Stderr, "[DEBUG] First 16 bytes after decompression (hex): %x\n", decompressed[:min(16, len(decompressed))])
					fmt.Fprintf(os.Stderr, "[DEBUG] Is valid UTF-8 after decompression: %v\n", utf8.Valid(decompressed))
					fmt.Fprintf(os.Stderr, "[DEBUG] Is likely text: %v\n", isLikelyText(decompressed))
				}
				value = decompressed
				wasDecompressed = true
			}
		}

		// If output file is specified, write to file
		if outputFile != "" {
			if err := os.WriteFile(outputFile, value, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Key: %s\n", key)
			fmt.Printf("Size: %d bytes\n", len(value))
			fmt.Printf("Output written to: %s\n", outputFile)
			if item.Flags != 0 {
				fmt.Printf("Flags: %d\n", item.Flags)
			}
			return
		}

		// Display metadata (unless value-only mode)
		if !valueOnly {
			fmt.Printf("Key: %s\n", key)
			if wasDecompressed {
				fmt.Printf("Size: %d bytes (decompressed from %d bytes)\n", len(value), originalSize)
			} else {
				fmt.Printf("Size: %d bytes\n", len(value))
			}
			if item.Flags != 0 {
				fmt.Printf("Flags: %d\n", item.Flags)
			}
			if item.Expiration != 0 {
				fmt.Printf("Expiration: %d\n", item.Expiration)
			}
			fmt.Printf("Data type: ")
			if isLikelyText(value) {
				fmt.Printf("text\n")
			} else {
				fmt.Printf("binary\n")
			}
			fmt.Println()
		}

		// Display the value based on format
		switch outputFormat {
		case "hex":
			if valueOnly {
				fmt.Printf("%s", hex.Dump(value))
			} else {
				fmt.Printf("Value (hex):\n%s\n", hex.Dump(value))
			}
		case "raw":
			if valueOnly {
				fmt.Printf("%s", string(value))
			} else {
				fmt.Printf("Value:\n%s\n", string(value))
			}
		default: // "auto"
			if isLikelyText(value) {
				if valueOnly {
					fmt.Printf("%s", string(value))
				} else {
					fmt.Printf("Value:\n%s\n", string(value))
				}
			} else {
				if !valueOnly {
					fmt.Printf("Value (binary data - %d bytes, showing hex):\n", len(value))
				}
				// Show first 512 bytes in hex dump
				displayLen := len(value)
				if displayLen > 512 {
					displayLen = 512
				}
				fmt.Printf("%s", hex.Dump(value[:displayLen]))
				if len(value) > 512 && !valueOnly {
					fmt.Printf("... (%d more bytes not shown, use --output-file to save full content)\n", len(value)-512)
				}
			}
		}
	},
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isGzipped checks if data appears to be gzip compressed
func isGzipped(data []byte) bool {
	return len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b
}

// isZlib checks if data appears to be zlib compressed
func isZlib(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	// Zlib compression starts with 0x78 followed by compression level indicator
	// 0x78 0x01 - No Compression/low
	// 0x78 0x9C - Default Compression
	// 0x78 0xDA - Best Compression
	return data[0] == 0x78 && (data[1] == 0x01 || data[1] == 0x5e || data[1] == 0x9c || data[1] == 0xda)
}

// tryGzipDecompress attempts to decompress gzip data
func tryGzipDecompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return decompressed, nil
}

// tryZlibDecompress attempts to decompress zlib data
func tryZlibDecompress(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return decompressed, nil
}

// isLikelyText checks if data is likely to be human-readable text
func isLikelyText(data []byte) bool {
	if len(data) == 0 {
		return true
	}

	// Check if it's valid UTF-8
	if !utf8.Valid(data) {
		return false
	}

	// Count printable characters
	printable := 0
	for _, b := range data {
		if (b >= 32 && b <= 126) || b == '\n' || b == '\r' || b == '\t' {
			printable++
		}
	}

	// If more than 95% is printable, consider it text
	return float64(printable)/float64(len(data)) > 0.95
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().StringVarP(&outputFormat, "format", "f", "auto", "Output format: auto, hex, or raw")
	getCmd.Flags().StringVarP(&outputFile, "output-file", "o", "", "Write value to file instead of stdout")
	getCmd.Flags().BoolVarP(&tryDecompress, "decompress", "d", false, "Try to decompress gzip data (auto-detects by default)")
	getCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose debugging information")
	getCmd.Flags().BoolVarP(&valueOnly, "value-only", "q", false, "Print only the value without metadata")
}
