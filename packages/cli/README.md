# Konflux Issues CLI

A command-line interface for managing Konflux issues, written in Go. This CLI tool provides a simple and efficient way to interact with the Konflux Issues API.

## Features

- List, filter, and search issues
- Get detailed information about specific issues
- Resovle issues manually
- Works as a standalone CLI and as a kubectl plugin
- Automatically detects and uses the current kubectl namespace
- Configurable API endpoint
- Colorized output for better readability
- Support for machine-readable output formats (JSON, YAML)

## Installation

### Option 1: Build from source

```bash
git clone https://github.com/konflux-ci/kite.git
cd packages/cli
make build
make install
```

### Setup as kubectl plugin

```bash
make kubectl-plugin
```

## Usage

### Standalone CLI

```bash
# List issues in a namespace
konflux-issues list -n team-alpha

# Filter issues by type
konflux-issues list -n team-alpha -t build

# Filter issues by severity
konflux-issues list -n team-alpha -s critical

# Get details for a specific issue
konflux-issues details -i <id> -n team-alpha

# Configure the API URL
konflux-issues config set-api-url http://localhost:8080/api/v1

# Show current configuration
konflux-issues config

# Reset configuration to defaults
konflux-issues config reset
```

### As a kubectl plugin

```bash
# List issues in current kubectl context namespace
kubectl issues list

# List issues in a specific namespace
kubectl issues list -n team-alpha

# Get details for an issue
kubectl issues details -i <id>

# Search for issues
kubectl issues search "dependency"

# Resolve an issue
kubectl issues resolve -i <id>
```

## Output Formats

The CLI supports three output formats:

1. **Table (default)**: Human-readable tabular format
2. **JSON**: Machine-readable JSON format
3. **YAML**: Machine-readable YAML format

Specify the output format using the `-o` or `--output` flag:

```bash
konflux-issues list -n team-alpha -o json
konflux-issues details -i failed-build-frontend -o yaml
```

## Configuration

The CLI uses a configuration file stored at `~/.konflux-issues/config.yaml`. You can modify settings using the `config` command or by directly editing this file.

Default configuration:
```yaml
api_url: http://localhost:8080/api/v1
```

You can also set the API URL using the `KONFLUX_API_URL` environment variable.

## Development

### Prerequisites

- Go 1.16 or later
- Make (optional)

### Building

```bash
# Build the binary
go build -o konflux-issues main.go

# Run tests
go test ./...

# Install locally
go install
```

## License

MIT
