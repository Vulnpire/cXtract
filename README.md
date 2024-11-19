
# cXtract

is designed to search cloud data for subdomains and IP addresses, with features like concurrent processing, subdomain, and IP extraction.

## Features
- Search cloud data for domains and subdomains.
- Extract IP addresses or subdomains with the `-only-ips` and `-subs` flags.
- Update cloud data from predefined sources.
- Supports concurrency for faster searches.
- Simple input handling through stdin (via `echo` or `cat`).
- Verbose logging with the `-v` flag for detailed output.

## Requirements
- Go 1.22 or higher.

## Installation

```go install -v github.com/Vulnpire/cXtract@latest```

## Usage

### Search for Subdomains
To search for subdomains of a domain and display them one per line:
```
echo "example.com" | cXtract -subs
```

### Search for IP Addresses Only
To search for IP addresses from cloud data:
```
echo "example.com" | cXtract -only-ips
```

### Set Concurrency Level
To specify the concurrency level for searches:
```bash
echo "example.com" | cXtract -c 10 -subs
```

## Flags
- `-duc` : Disable updates of cloud data.
- `-v` : Enable verbose output.
- `-c <level>` : Set the concurrency level.
- `-subs` : Extract and display subdomains only.
- `-only-ips` : Extract and display IP addresses only.
