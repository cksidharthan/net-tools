# net-tools

NOTE: this is a work in progress repo that I'm using to test something that I'm building.

A simple REST API for running common network tools like `ping`, `traceroute`, `nslookup`, etc.

## Features

- WebSocket-based real-time network diagnostics
- Currently supports:
  - Ping with configurable parameters

## Quick Start

1. Install and run:

```bash
git clone https://github.com/cksidharthan/net-tools.git
cd net-tools
go mod download
go run main.go
```

## API Usage

### Ping
Connect to `ws://localhost:3000/ping` and send:

```json
{
  "address": "example.com",
  "count": 5,
  "wait": 1,
  "packet_size": 56,
  "timeout": 5
}
```

## Development

Built with:
- Go Chi router
- Gorilla WebSocket
- Air for hot reloading

Run development server:
```bash
task run
```