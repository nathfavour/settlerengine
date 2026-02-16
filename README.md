# SettlerEngine

The Agentic Settlement Gateway for AI Agents and Human Users.

## Features
- **x402 Protocol**: Built-in support for HTTP 402 "Payment Required" flows.
- **EIP-712 Verification**: Cryptographically secure payment intent verification.
- **Multi-Chain**: Support for Base, Cronos, Avalanche, and Polygon.
- **Stateless**: High-throughput verification without database overhead.

## Deployment

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
