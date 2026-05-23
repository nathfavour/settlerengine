# Roadmap: Automated Yield Settlement Layer (Riquid Integration)

This roadmap outlines the integration of the Riquid Self-Driving Yield Engine into the SettlerEngine hexagonal architecture.

## Phase 1: Domain Modeling & Port Definitions (`core/domain`) 🏗️
- [x] **Define `YieldStrategy` Entity**
- [x] **Define `YieldProvider` Port (Interface)**
- [x] **Update `SettlementEngine` Port:** Add `DepositToYield` and `WithdrawFromYield` methods.

## Phase 2: BSC Infrastructure & Adapters (`pkg/chains`) ⛓️
- [x] **BSC RPC Provider:** Implement a Geth-compatible provider for BNB Smart Chain (BSC).
- [x] **Asset Support:** Add configuration and tracking for BNB, USDT (BEP-20), and BUSD.
- [x] **Contract Bindings:** Generate Go ABIs/bindings for Riquid Vaults.

## Phase 3: Riquid Driven Adapter (`pkg/yield`) 💸
- [x] **`riquid_adapter.go`:** Implementation of the `YieldProvider` interface for Riquid Yield Engine using generated bindings.
- [x] **Withdrawal Logic:** Implement `WithdrawFromYield` method.
- [x] **State Machine Integration:** Logic to encode/decode calls to Riquid strategy contracts via bindings.

## Phase 4: Self-Driving Yield Automation (`core/domain/service`) 🤖
- [x] **Auto-Route Service:** Implementation of routing logic upon `SETTLEMENT_CONFIRMED` events.
- [x] **Threshold Logic:** Implement gas-efficiency triggers to prevent micro-transactions.
- [x] **Cron Worker:** Develop a "Self-Driving" background worker for periodic harvesting and reinvestment.
- [x] **Event Bus Wiring:** Settlement events are now published via `LocalBus` and consumed by `YieldService`.

## Phase 5: Account Abstraction & Session Keys (`pkg/crypto` & `pkg/yield`) 🔐
- [x] **ERC-4337 Integration:** Logic to manage funds via non-custodial account abstraction (Provider + UserOps).
- [x] **Session Key Manager:** Sign "Harvest" and "Reinvest" transactions using restricted-scope keys.
- [x] **Paymaster Integration:** Support for gas sponsorship on BSC via `Paymaster` client.

## Phase 6: Observability & Validation (`pkg/metrics`) 📊
- [x] **Prometheus Metrics:** Track APY performance, total value locked (TVL) in yield, and "Time-to-Settle".
- [x] **Integration Tests & Wiring:** End-to-end wiring in `settlerd`, unit tests for yield logic, and event-driven routing tests.

---
**Core Requirement:** All implementations must maintain the **Non-Custodial** nature of SettlerEngine. Automation must be achieved through cryptographic delegation (Session Keys), not centralized management.
