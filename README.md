# SettlerEngine

**The protocol-agnostic settlement gateway for the agentic era.**

SettlerEngine is a high-performance reverse proxy and facilitator designed to bridge AI agents and digital resources through autonomous payments. It implements the x402 protocol, providing a cryptographically secure, stateless, and persistent way to handle machine-to-machine commerce across multiple EVM-compatible chains. Built with sovereign environments in mind, it supports local communication via Unix Domain Sockets and seamless distribution through Anyisland.

### Product Description

In the emerging agentic economy, AI agents require the ability to discover, negotiate, and pay for resources—API access, compute power, or proprietary data—without human intervention. SettlerEngine acts as the "Stripe for Machines," providing the infrastructure for this autonomous commerce.

At its core, SettlerEngine intercepts unauthenticated requests and issues cryptographic challenges using the **x402 Protocol**. Agents respond by signing an **EIP-712 Intent-to-Pay** message, which the engine verifies against blockchain state. By focusing on **stateless verification** and leveraging cryptographic proofs, SettlerEngine achieves the high throughput necessary for high-frequency agentic interactions while maintaining absolute financial security.

Whether running as a global gateway or a local sidecar for co-located agents, SettlerEngine ensures that every digital interaction is backed by a valid settlement intent, unlocking the full potential of autonomous digital markets.

---

## 🚀 Key Features

- **x402 Handshake**: Native support for HTTP 402 "Payment Required" flows optimized for AI Agent parsing.
- **EIP-712 Verification**: Secure, typed signature verification ensuring agents only pay exactly what they intended.
- **Multi-Chain Native**: Out-of-the-box support for **Base**, **Cronos**, **Avalanche**, and **Polygon**.
- **Stateless Authorization**: High-speed verification loop that minimizes database latency.
- **Persistent Idempotency**: CGO-free SQLite backend ensures verified payments are cached across restarts.
- **Local-First (UDS)**: Secure Unix Domain Socket support for ultra-low latency communication between local processes.
- **Anyisland Ready**: Built-in "Pulse" awareness and auto-registration for the Anyisland sovereign ecosystem.

---

## 📦 Deployment

### One Binary Build
SettlerEngine compiles into a single, dependency-free binary containing both the proxy and facilitator.
```bash
CGO_ENABLED=0 go build -o settler ./cmd/settler
```

### Quick Start (Proxy Mode)
Start a reverse proxy that requires a $1.00 USDC payment on Base Sepolia before forwarding traffic:
```bash
./settler proxy -target http://your-api:8081 -amount 1000000 -chain-id 84532
```

### Docker & Podman
Build and run using the provided multi-stage Dockerfile:
```bash
docker build -t settler-engine .
docker run -p 8080:8080 settler-engine proxy -target http://host.docker.internal:8081
```

### Anyisland
If you use [Anyisland](https://github.com/anyisland), the engine is fully managed:
```bash
anyisland install github.com/nathfavour/settlerengine
```

---

## 🔌 Local Integration

SettlerEngine is designed to be a robust local partner for agents running on the same host.

### Data Directory
The engine respects OS standards for data storage:
- **Linux**: `~/.config/settlerengine/`
- **macOS**: `~/Library/Application Support/settlerengine/`
- **Windows**: `%AppData%\settlerengine\`

### Unix Domain Socket (UDS)
Local agents can bypass the network stack by connecting to the socket at `settler.sock` within the data directory. This provides a secure, zero-overhead channel for payment processing and status checks.

---

## 🛠️ Architecture

SettlerEngine follows **Hexagonal Architecture** principles:
- **Core**: Pure domain logic for Invoices and Money.
- **Pkg**: Reusable adapters for Crypto, Storage (SQLite), and UDS.
- **Cmd**: Unified entry point for all engine sub-commands.

For a deep dive, see the [Architecture Documentation](./docs/docs/architecture.md).

---

## 📖 Documentation

- [Agent Integration Guide](./docs/docs/agents.md)
- [x402 Protocol Deep Dive](./docs/docs/x402.md)
- [Local Integration & UDS](./docs/docs/local-integration.md)
- [Chain Configurations](./docs/docs/chains.md)

---

## 📜 License
MIT License. See [LICENSE](LICENSE) for details.
