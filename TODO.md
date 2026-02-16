# SettlerEngine Roadmap: The Agentic Settlement Era

This document outlines the path toward a high-throughput, protocol-agnostic settlement gateway for AI agents and human users.

## Phase 1: Foundation & Workspace 🏗️
- [x] Initialize Go Workspace (`go.work`)
- [x] Define Hexagonal Architecture boundaries
- [x] Establish `core`, `pkg`, and `apps` module separation
- [x] Basic `Money` value object and `Invoice` aggregate

## Phase 2: x402 Protocol Implementation (`pkg/x402`) 💸
- [ ] **PaymentDescriptor:** Define standard JSON structures for 402 responses (amount, asset, network, recipient).
- [ ] **State Machine:** Implement the HTTP 402 "Payment Required" lifecycle.
- [ ] **Header Parser:** Logic to extract and decode `X-PAYMENT` or `Payment-Signature` headers.
- [ ] **Nonce Manager:** Cryptographic challenge (session ID) generator to prevent replay attacks.

## Phase 3: Technical Heart - Verification (`pkg/crypto`) 🔐
- [ ] **EIP-712 Implementation:** Define the SettlerEngine Domain Separator and standard signature format.
- [ ] **Signature Recovery:** Implementation of signature recovery to verify agent wallet addresses.
- [ ] **Stateless Authorization:** Logic to verify "Intent to Pay" without heavy database dependencies.

## Phase 4: Chain Observer & Settlement (`pkg/chains`) ⛓️
- [ ] **Unified RPC Multi-Client:** Support for Base, Cronos, and SKALE.
- [ ] **Non-Custodial Sweeper:** Logic to ensure funds route directly to merchants.
- [ ] **Transaction Verifier:** Confirm on-chain transfers match the issued challenges.

## Phase 5: The "Qualifying" Flow (MVP) 🚀
- [ ] **settler-proxy:** Implement the x402 Reverse Proxy Gateway.
  - [ ] Intercept unauthenticated requests.
  - [ ] Issue 402 Challenges.
  - [ ] Proxy authorized requests to backend services.
- [ ] **settlerd:** The Facilitator Daemon.
  - [ ] Manage long-lived settlement state.
  - [ ] Broadcast/Verify on-chain transfers.

## Phase 6: Demos & Examples 🛠️
- [ ] Example agent implementation (Python/TS) paying for an API.
- [ ] Example merchant integration.

---
**Core Principle:** Focus on **Statelessness**. Rely on the blockchain as the source of truth and cryptographic signatures as proof of intent.
