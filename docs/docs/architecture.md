# Architecture

SettlerEngine follows a **Hexagonal Architecture** (Ports and Adapters) to ensure the domain logic remains decoupled from infrastructure.

## Modules

- **Core:** Contains the domain models and business services.
- **Pkg:** Shared utilities including `x402` protocol logic and `crypto` verification.
- **Apps:** Entry points for the engine daemon (`settlerd`) and the gateway proxy.
