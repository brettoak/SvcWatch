# SvcWatch

SvcWatch is a lightweight, real-time Nginx log monitoring system written in Go.

## Overview

SvcWatch is designed to monitor Nginx access logs in real-time. It acts as an intelligent agent that "tails" your log files, parses them on-the-fly into structured data, and exposes statistics via a RESTful API.

It is built with a modular architecture, making it easy to extend for different storage backends (Redis, Kafka) or logic requirements.

## Features

- **Real-time Tailing**: Continuously watches log files for new entries, handling file rotation (logrotate) automatically.
- **Regex Parsing**: Robust parsing of the default Nginx access log format.
- **Thread-Safe Storage**: Implements concurrency-safe in-memory storage for high-throughput environments.
- **REST API**: Provides HTTP endpoints to query system status and log statistics (built with [Gin](https://github.com/gin-gonic/gin)).

## Architecture

The project is organized into modular components:

- **Collector** (`internal/collector`): Responsible for file I/O and tailing.
- **Parser** (`internal/parser`): Transforms raw log lines into structured `LogEntry` objects.
- **Storage** (`internal/storage`): Interface-driven storage layer. Currently supports in-memory storage.
- **Monitor** (`nginx-log-monitor/monitor.go`): The facade that orchestrates these components.

## Getting Started

### Prerequisites

- Go 1.22 or higher.

### Installation

```bash
# Clone the repository
git clone https://github.com/your-repo/SvcWatch.git
cd SvcWatch

# Download dependencies
go mod tidy
```

### Usage

1. **Prepare a Log File**  
   The application monitors `./access.log` by default. You can create a dummy file to test:
   ```bash
   echo '127.0.0.1 - - [24/Jan/2026:15:00:00 +1100] "GET / HTTP/1.1" 200 612 "-" "-"' >> access.log
   ```

2. **Run the Application**
   ```bash
   go run main.go
   ```
   *Note: If you encounter permission issues with dependencies, try `export GOMODCACHE=/tmp/gomodcache` before running.*

3. **Check Status**
   The server listens on `localhost:8080`.
   
   - **Health Check**:
     ```bash
     curl http://localhost:8080/ping
     # Output: {"message":"pong"}
     ```
   
   - **Log Statistics**:
     ```bash
     curl http://localhost:8080/stats
     # Output: {"total_logs": 1}
     ```

## Project Structure

```text
SvcWatch/
├── main.go                 # Entry point
├── go.mod                  # Root module definition
├── nginx-log-monitor/      # Core logic module
│   ├── monitor.go          # Public Façade
│   ├── config/             # Configuration
│   └── internal/           # Private internal components
│       ├── collector/      # Log tailing implementation
│       ├── parser/         # Log parsing logic
│       ├── storage/        # Data storage layers (Memory, etc.)
│       ├── model/          # Data models
│       └── service/        # Business logic
└── README.md
```

## Roadmap

- [x] Basic Log Tailing & Parsing
- [x] In-Memory Storage
- [ ] Redis Integration for Persistence
- [ ] Kafka Producer for Log Streaming
- [ ] Advanced Layout/Metrics (QPS, Latency P99)
- [ ] Web Dashboard
