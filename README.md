This is a project for URL shortener

## Setup & Installation

### Prerequisites
- Go 1.26.1 or higher
- k6 (for load testing)

### Installing k6

**macOS:**
```bash
brew install k6
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install k6
```

**Windows:**
```bash
choco install k6
```

Or download from: https://k6.io/docs/getting-started/installation/

### Running the Server

1. Navigate to the server directory:
   ```bash
   cd server
   ```

2. Download dependencies (optional - they auto-download on first run):
   ```bash
   go mod download
   ```

3. Run the server:
   ```bash
   go run ./cmd/api/main.go
   ```

The dependencies are automatically managed by `go.mod` and `go.sum`, so you only need Go installed on any machine.

### Running the Client

```bash
cd client
go run ./src/main.go
```

### Load Testing

Make sure the server is running first, then:

```bash
cd load-test
k6 run post_req.js  # Test POST requests
k6 run get_req.js   # Test GET requests
```
