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

## 🏆 Mantle Turing Test Hackathon 2026 Integration

SettlerEngine has been extended to natively support the **Mantle Network** as part of the Turing Test Hackathon, bridging high-speed settlement with on-chain agent trust and verifiable reputation.

### ⚓ On-Chain Logging (Mantle Sepolia)
A **Logging & Settlement Registry** is deployed to Mantle Sepolia to provide an immutable, verifiable footprint of all agent-to-agent transactions. This satisfies the requirement for recording key agentic decisions on-chain.

-   **SettlerRegistry Address:** `0x33aE8331a2406EEc3A33483001aC5650DA2e0662`
-   **Network:** Mantle Sepolia (Chain ID: 5003)
-   **Functionality:** Anchors `AgentPaymentLogged` events including Agent IDs, Invoice IDs, and transaction metadata.

### 🤖 ERC-8004: Trustless Agent Identity
The gateway now implements the **ERC-8004** standard for decentralized agent identity and reputation. 
-   **Identity Verification:** Intercepts x402 payments and validates the signer against the ERC-8004 Identity Registry.
-   **Automated Reputation:** Upon successful payment settlement, the engine automatically posts positive trust signals to the Mantle-based Reputation Registry, closing the loop between financial action and trust.

### 🏗️ CI/CD & Cloud Infrastructure
To facilitate rapid iteration and verifiable builds, the project utilizes:
-   **GitHub Container Registry (GHCR):** Automated builds published to `ghcr.io/nathfavour/settlerengine`.
-   **Smart Pipelines:** GitHub Actions-driven CI/CD with layer caching and a `BUILD_ENABLED` toggle for cost-efficient iterations.

### 🛠️ Hackathon Commands
Execute a policy-protected payment on Mantle or other supported nets:
```bash
./settler pay -to <RECIPIENT> -amount <WEI> -key <AGENT_PRIVATE_KEY>
```

Run the full agentic demo (identity resolution -> policy check -> on-chain anchoring):
```bash
./settler demo
```

### 🔗 Proof of On-chain Activity (Mantle Sepolia)
As proof of native integration and live agentic activity, several demo transactions have been anchored to Mantle Sepolia:

-   **Agent Payment Logged:** [`0x7bfb4b8b93b6e4bbbc27248e3fb2d4ce7c5240f5eb8da8af57f992c1a1c8ac72`](https://explorer.sepolia.mantle.xyz/tx/0x7bfb4b8b93b6e4bbbc27248e3fb2d4ce7c5240f5eb8da8af57f992c1a1c8ac72)
-   **SettlerRegistry Deployment:** [`0x33aE8331a2406EEc3A33483001aC5650DA2e0662`](https://explorer.sepolia.mantle.xyz/address/0x33aE8331a2406EEc3A33483001aC5650DA2e0662)

---
