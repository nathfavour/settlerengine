# SettlerEngine

The Agentic Settlement Gateway for AI Agents and Human Users.

## Features
- **x402 Protocol**: Built-in support for HTTP 402 "Payment Required" flows.
- **EIP-712 Verification**: Cryptographically secure payment intent verification.
- **Multi-Chain**: Support for Base, Cronos, Avalanche, and Polygon.
- **Stateless**: High-throughput verification without database overhead.

## Deployment

### Local Integration (UDS & SQLite)
SettlerEngine is designed for robust local agent communication using Unix Domain Sockets (UDS) and persistent storage via CGO-free SQLite.

#### Data Directory
By default, the engine uses the standard system configuration directory:
- **Linux**: `~/.config/settlerengine/`
- **macOS**: `~/Library/Application Support/settlerengine/`
- **Windows**: `%AppData%\settlerengine\`

This directory contains:
- `settler.db`: SQLite database for persistent payment verification.
- `settler.sock`: Unix Domain Socket for local process communication.

#### Connecting via UDS
Local agents can connect to the engine using the socket file. This is ideal for services running on the same machine that need to verify payments or request signatures without network overhead.

Example (Go):
```go
conn, err := net.Dial("unix", "~/.config/settlerengine/settler.sock")
```

#### SQLite Persistence
Verification states are stored in a CGO-free SQLite database. This ensures that even after a restart, previously verified payment intents remain valid, preventing redundant on-chain checks or 402 challenges.

### One Binary Build
You can build the `settler` binary which contains both the proxy and the facilitator.
```bash
go build -o settler ./cmd/settler
```

### Running the Proxy
```bash
./settler proxy -target http://your-api:8080 -listen :8080
```

### Docker
Build the image:
```bash
docker build -t settler-engine .
```

Run with Docker Compose:
```bash
docker-compose up
```

### Podman
Podman is fully supported as an alternative to Docker.
```bash
podman build -t settler-engine .
podman run -p 8080:8080 settler-engine proxy -target http://localhost:8081
```

## Development
- `pkg/x402`: Middleware and protocol logic.
- `pkg/crypto`: EIP-712 and signature verification.
- `cmd/settler`: Main entry point.
