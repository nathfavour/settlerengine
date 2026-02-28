# SettlerEngine: The Agentic BTCPay Server Roadmap

SettlerEngine aims to be the sovereign settlement layer for the AI economy. While it excels at machine-to-machine (x402) flows, it requires several improvements to match the robustness of BTCPay Server while leading the agentic era.

## 🛠️ Shortcomings & Missing Features

### 1. Protocol & Chain Support
- [ ] **Non-EVM Support:** Implement Solana (SPL) and Bitcoin (Lightning/L402) providers as outlined in `ARCHITECTURE.md`.
- [ ] **Stablecoin Focus:** Deepen integration with USDC/USDT across all supported chains with auto-swaps to yield-bearing assets.

### 2. Merchant & Human Experience
- [ ] **Headless Dashboard:** A Next.js/React UI for merchants to visualize:
    - [ ] Real-time settlement stream.
    - [ ] Yield performance (APY/TVL).
    - [ ] Agent interaction logs (x402 success/failure rates).
- [ ] **Multi-Tenant Stores:** Support for multiple "stores" or "projects" under a single SettlerEngine instance.
- [ ] **Webhook System:** Standardized webhooks for notifying legacy backends of successful settlements.

### 3. Agent-Centric Features
- [ ] **Dynamic Pricing Engine:** Allow merchants to define pricing logic (e.g., "discount for agents with high reputation" or "premium for high-bandwidth requests").
- [ ] **Agent Sidecar:** A companion client that agents can run to "auto-solve" x402 challenges using a pre-funded local wallet.
- [ ] **SLA Verification:** Cryptographic proof that the digital resource was delivered *after* payment (Optimistic Settlement).

### 4. Technical Debt & Implementation Gaps
- [x] **Event Bus:** Integrate Watermill for decoupled communication between `SettlementEngine` and `YieldService`. (Moved from Architecture)
- [ ] **Robust ERC-4337:** Complete the Account Abstraction provider for gasless automation and restricted session keys.
- [ ] **Yield Withdrawals:** Implement the `WithdrawFromYield` method in the `RiquidAdapter`.
- [ ] **State Persistence:** Ensure the SQLite backend tracks Yield TVL and historical earnings accurately.

## ✅ Completed Tasks (Cleaned up from previous phases)
- [x] x402 Middleware for HTTP 402 Handshakes.
- [x] EIP-712 Intent-to-Pay verification.
- [x] Hexagonal Core (Invoice, Money, Settlement domains).
- [x] Basic Riquid Yield Adapter skeleton.
- [x] Background Harvesting Worker.
- [x] Multi-chain RPC Client wrapper.
