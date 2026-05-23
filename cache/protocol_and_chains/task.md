# Protocol & Multi-Chain Native Support

This area of concentration deals with expanding SettlerEngine's multi-chain architecture from standard EVM to support non-EVM and UTXO protocols natively, creating a truly protocol-agnostic settlement layer for agents and humans.

## Milestones & Goals

### 1. Robust Monero (XMR) Private View-Key Integration
*   Implement subaddress derivation locally or via `monero-wallet-rpc` using the merchant's **Private View Key**.
*   Scan incoming transactions without exposing the **Private Spend Key** to preserve complete non-custodial security.
*   Manage invoice status transitions based on block confirmations.

### 2. UTXO Production Suite (BTC & LTC)
*   Integrate direct JSON-RPC interface with `bitcoind` and `litecoind` nodes for address-pool reservation and raw transaction monitoring.
*   Establish LND gRPC stream integrations to detect Lightning payments in real-time.
*   Support **L402 (LSAT)** protocol flow for high-frequency agent micro-payments.

### 3. Solana SPL & Tron TRC-20 Account Trackers
*   Implement Solana websocket connection to monitor SPL token transfer instructions (SOL, USDC, USDT) on specific addresses.
*   Parse Tron TRC-20 log structures using native Go Protobuf decoders over public RPC gateways.
