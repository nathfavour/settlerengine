# Automated Yield Generation & Cryptographic Delegation

To maximize treasury utilization for self-hosters and businesses, SettlerEngine utilizes a "self-driving" yield service. Crucially, all automation must retain standard **Non-Custodial** characteristics.

## Milestones & Goals

### 1. Riquid Vault Autopilot
*   Fully connect the domain to concrete BSC smart contract wrappers (`riquid_adapter.go`) using Go ABI bindings.
*   Implement periodic cron-based harvesting and reinvesting loops for optimal gas efficiency.
*   Enforce a deposit sweep threshold (e.g. minimum deposit equivalent of $10 in native/stablecoin) to prevent micro-transaction gas waste.

### 2. ERC-4337 Account Abstraction (AA) Integration
*   Build a non-custodial smart wallet framework using ERC-4337 UserOperations.
*   Integrate bundler RPC connections to submit UserOps directly.
*   Implement Paymaster clients to allow gas-free automation or sponsorship of harvesting transactions on BSC and layer-2 EVMs.

### 3. Secure Session Key Manager
*   Implement cryptographic session keys that delegate restricted spending permissions (e.g. "only allowed to call vault harvest() or vault deposit() up to $X00") to local nodes.
*   Store session keys in encrypted keystore files.
