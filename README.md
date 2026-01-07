# mcache

A simple, stable CLI tool for interacting with remote memcached servers.

## Features

- Get values from memcached keys
- Auto-detect and decompress gzip and zlib-compressed data
- Handle both text and binary data intelligently
- Multiple output formats (auto, hex, raw)
- Value-only mode for piping to other commands
- Save values to files
- Display key metadata (flags, expiration, size)
- Flexible server configuration
- Single binary with no runtime dependencies
- Clear error messages

## Installation

### From Source

```bash
git clone <repository-url>
cd mcache
go build -o mcache
```

### Install to System

```bash
go install
```

Or copy the binary to your PATH:

```bash
cp mcache /usr/local/bin/
```

## Usage

### Basic Commands

```bash
# Get a key from localhost:11211 (default)
mcache get mykey

# Specify custom server and port
mcache get mykey --server=192.168.1.100 --port=11211

# Using short flags
mcache get mykey -s cache.example.com -p 11211
```

### Global Flags

All commands support these global flags:

- `-s, --server` - Memcached server address (default: "localhost")
- `-p, --port` - Memcached server port (default: 11211)
- `-h, --help` - Show help information

## Commands

### get

Retrieve and display the contents of a specific key.

```bash
mcache get <key> [flags]
```

**Flags:**
- `-f, --format` - Output format: `auto` (default), `hex`, or `raw`
- `-o, --output-file` - Write value to file instead of stdout
- `-d, --decompress` - Force decompression (auto-detects by default)
- `-q, --value-only` - Print only the value without metadata (useful for piping)
- `-v, --verbose` - Show verbose debugging information

**Behavior:**
- Automatically detects and decompresses gzip and zlib-compressed data
- For text data: displays as readable text
- For binary data: shows hex dump (first 512 bytes)
- Use `--output-file` to save full binary data to a file
- Use `--value-only` to output just the value for piping to other commands

**Example Output (text):**
```
Key: mykey
Size: 24 bytes
Flags: 0

Value:
cached_value_here
```

**Example Output (binary):**
```
Key: allcfcerts
Size: 45678 bytes
Flags: 3

Value (binary data - 45678 bytes, showing hex):
00000000  1f 8b 08 00 00 00 00 00  00 ff ec 7d 5b 6f 24 c7  |...........}[o$.|
00000010  75 9e 01 bf bf 40 2f 20  80 00 2c 4e 65 66 5d 39  |u....@/ ..,Nef]9|
...
```

**Error Cases:**
- Key not found: Returns error message and exit code 1
- Connection failure: Returns error with server address and exit code 1

### set

Set a key-value pair in the memcached server.

```bash
mcache set <key> [value] [flags]
```

**Flags:**
- `-e, --expiration` - Expiration time in seconds (0 = never expires, default: 0)
- `-F, --flags` - Flags to store with the item (default: 0)
- `-i, --input-file` - Read value from file

**Value Sources (in priority order):**
1. Command argument: `mcache set mykey "my value"`
2. File: `mcache set mykey --input-file=data.txt`
3. Stdin: `echo "my value" | mcache set mykey`

**Example Output:**
```
Set key 'mykey' (24 bytes)
```

## Examples

```bash
# Get a text value
mcache get session:user:12345

# Get binary/compressed data (auto-decompresses gzip and zlib)
mcache get allcfcerts

# Save binary data to a file
mcache get image:profile:avatar -o avatar.jpg

# View data as hex dump
mcache get encrypted:token --format=hex

# Force raw output (skip auto-detection)
mcache get mykey --format=raw

# Get only the value (useful for piping)
mcache get mykey --value-only
mcache get mykey -q

# Pipe JSON value to jq for formatting
mcache get config:settings -q | jq .

# Count lines in cached text
mcache get large:text -q | wc -l

# Get from production server
mcache get homepage_cache -s prod-cache-01.internal -p 11211

# Get from remote server with custom port
mcache get api:response:xyz --server=10.0.1.50 --port=11212

# Set a simple value
mcache set mykey "hello world"

# Set with expiration (1 hour)
mcache set session:token:abc "user123" --expiration=3600

# Set from a file
mcache set config:settings --input-file=settings.json

# Set from stdin (pipe)
echo '{"status": "ok"}' | mcache set api:health

# Set on a specific server
mcache set cache:data "value" -s cache.example.com -p 11211
```

## Building

Requirements:
- Go 1.16 or later

```bash
# Build for current platform
go build -o mcache

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o mcache-linux-amd64
```

## Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [gomemcache](https://github.com/bradfitz/gomemcache) - Memcached client

## Architecture

- **Language**: Go
- **CLI Framework**: Cobra for clean command structure
- **Memcached Client**: bradfitz/gomemcache (stable, maintained library)

Design principles:
- Single binary distribution
- Minimal dependencies
- Clear error handling
- Extensible command structure

## Future Commands

Planned additions:
- `delete` - Delete a key
- `stats` - Show server statistics
- `flush` - Flush all keys

## License

[Choose your license]

## Contributing

Contributions welcome! Please feel free to submit a Pull Request.
