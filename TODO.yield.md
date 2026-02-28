# Roadmap: Automated Yield Settlement Layer (Riquid Integration)

This roadmap outlines the integration of the Riquid Self-Driving Yield Engine into the SettlerEngine hexagonal architecture.

## Phase 1: Domain Modeling & Port Definitions (`core/domain`) 🏗️
- [ ] **Define `YieldStrategy` Entity:**
    ```go
    type YieldStrategy struct {
        ID            string
        Provider      string // e.g., "Riquid"
        TargetAPY     float64
        AutoHarvest   bool
        LastHarvested time.Time
        VaultAddress  string
    }
    ```
- [ ] **Define `YieldProvider` Port (Interface):**
    ```go
    type YieldProvider interface {
        DepositToYield(ctx context.Context, amount money.Money, strategy YieldStrategy) error
        WithdrawFromYield(ctx context.Context, amount money.Money, strategy YieldStrategy) error
        GetAPY(ctx context.Context, vaultAddress string) (float64, error)
        Harvest(ctx context.Context, strategy YieldStrategy) error
    }
    ```
- [ ] **Update `SettlementEngine` Port:** Add `DepositToYield` and `WithdrawFromYield` methods.

## Phase 2: BSC Infrastructure & Adapters (`pkg/chains`) ⛓️
- [x] **BSC RPC Provider:** Implement a Geth-compatible provider for BNB Smart Chain (BSC).
- [/] **Asset Support:** Add configuration and tracking for BNB, USDT (BEP-20), and BUSD.
- [ ] **Contract Bindings:** Generate Go ABIs/bindings for Riquid Vaults and AsterDEX Earn contracts.

## Phase 3: Riquid Driven Adapter (`pkg/yield`) 💸
- [ ] **`riquid_adapter.go`:** Implementation of the `YieldProvider` interface for Riquid Yield Engine.
- [ ] **State Machine Integration:** Logic to encode/decode calls to Riquid strategy contracts.

## Phase 4: Self-Driving Yield Automation (`core/domain/service`) 🤖
- [ ] **Auto-Route Service:** Implementation of routing logic upon `SETTLEMENT_CONFIRMED` events.
- [ ] **Threshold Logic:** Implement gas-efficiency triggers to prevent micro-transactions.
- [ ] **Cron Worker:** Develop a "Self-Driving" background worker for periodic harvesting and reinvestment.

## Phase 5: Account Abstraction & Session Keys (`pkg/crypto` & `pkg/yield`) 🔐
- [ ] **ERC-4337 Integration:** Logic to manage funds via non-custodial account abstraction.
- [ ] **Session Key Manager:** Sign "Harvest" and "Reinvest" transactions using restricted-scope keys.
- [ ] **Paymaster Integration:** Support for gas sponsorship on BSC via `Paymaster` client.

## Phase 6: Observability & Validation (`pkg/metrics`) 📊
- [ ] **Prometheus Metrics:** Track APY performance, total value locked (TVL) in yield, and "Time-to-Settle".
- [ ] **Integration Tests:** End-to-end validation of the BSC -> Riquid flow on Testnet (BSC Sepolia).

---
**Core Requirement:** All implementations must maintain the **Non-Custodial** nature of SettlerEngine. Automation must be achieved through cryptographic delegation (Session Keys), not centralized management.
