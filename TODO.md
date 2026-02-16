# SettlerEngine Roadmap: The Agentic Settlement Era

This document outlines the path toward a high-throughput, protocol-agnostic settlement gateway for AI agents and human users.

## Phase 1: Foundation & Workspace 🏗️
- [x] Initialize Go Workspace (`go.work`)
- [x] Define Hexagonal Architecture boundaries
- [x] Establish `core`, `pkg`, and `apps` module separation
- [x] Basic `Money` value object and `Invoice` aggregate

## Phase 2: x402 Protocol Implementation (`pkg/x402`) 💸
- [x] **PaymentDescriptor:** Define standard JSON structures for 402 responses (amount, asset, network, recipient).
- [x] **State Machine:** Implement the HTTP 402 "Payment Required" lifecycle.
- [x] **Header Parser:** Logic to extract and decode `X-PAYMENT` or `Payment-Signature` headers.
- [x] **Nonce Manager:** Cryptographic challenge (session ID) generator to prevent replay attacks.

## Phase 3: Technical Heart - Verification (`pkg/crypto`) 🔐
- [x] **EIP-712 Implementation:** Define the SettlerEngine Domain Separator and standard signature format.
- [x] **Signature Recovery:** Implementation of signature recovery to verify agent wallet addresses.
- [x] **Stateless Authorization:** Logic to verify "Intent to Pay" without heavy database dependencies.

## Phase 4: Chain Observer & Settlement (`pkg/chains`) ⛓️
- [x] **Unified RPC Multi-Client:** Support for Base, Cronos, SKALE, Avalanche, and Polygon.
- [ ] **RPC Redundancy:** Implement fallback mechanism (dRPC/Thirdweb) for rate-limiting protection.
- [x] **Chain Configuration:** Mapping of ChainID to x402 Facilitator/USDC contract addresses.
- [ ] **EIP-3009 Integration:** Implement `transferWithAuthorization` support for gasless USDC transfers.
  - [ ] Specific verification against Base Sepolia (`0x036CbD...`) and Cronos zkEVM (`0xaa5b8...`).
- [ ] **Non-Custodial Sweeper:** Logic to ensure funds route directly to merchants.
- [ ] **Transaction Verifier:** Confirm on-chain transfers match the issued challenges.

## Phase 5: The "Qualifying" Flow (MVP) 🚀
- [x] **settler-proxy:** Implement the x402 Reverse Proxy Gateway.
  - [x] Intercept unauthenticated requests.
  - [x] Issue 402 Challenges.
  - [ ] Support for EIP-3009 payload injection.
  - [x] Proxy authorized requests to backend services.
- [ ] **settlerd:** The Facilitator Daemon.
  - [ ] Manage long-lived settlement state.
  - [ ] Multi-chain RPC management.
  - [ ] Broadcast/Verify on-chain transfers.

## Phase 6: Demos & Examples 🛠️
- [ ] Example agent implementation (Python/TS) paying for an API.
- [ ] Example merchant integration.

---
**Core Principle:** Focus on **Statelessness**. Rely on the blockchain as the source of truth and cryptographic signatures as proof of intent.
