# Architecture

SettlerEngine follows a **Hexagonal Architecture** (Ports and Adapters) to ensure the domain logic remains decoupled from infrastructure.

## Modules

- **Core:** Contains the domain models and business services.
- **Pkg:** Shared utilities including `x402` protocol logic, `crypto` verification, `storage` (SQLite), and `uds` (Unix Domain Sockets).
- **Apps/Cmd:** Entry points for the engine binary (`cmd/settler`), which manages the proxy and facilitator.

## Infrastructure Adapters

### Driving Adapters (Inputs)
- **HTTP Proxy:** Intercepts external agent requests.
- **UDS Server:** Provides a local channel for co-located agents.

### Driven Adapters (Outputs)
- **SQLite Storage:** Persists verified intents and session state.
- **Blockchain Multi-Client:** Communicates with RPC providers (Base, Polygon, etc.) for final settlement verification.

