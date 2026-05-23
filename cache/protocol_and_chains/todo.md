# Todo Checklist: Protocol & Chains

## Short-Term Goals (1-2 Weeks)
- [ ] Implement local Monero subaddress derivation using ed25519 cryptography.
- [ ] Scaffold standard `bitcoind` RPC client with authentication.
- [ ] Setup Solana RPC JSON client to query address balances.

## Medium-Term Goals (1 Month)
- [ ] Connect to `monero-wallet-rpc` for scanning view-key transactions.
- [ ] Integrate LND gRPC client and monitor invoice status streams.
- [ ] Parse Tron TRC-20 hex transfer event logs dynamically.
- [ ] Implement auto-recovery for RPC connection drops.

## Long-Term Goals (Production Gating)
- [ ] Support L402 / LSAT macaroon gating verification middleware.
- [ ] Optimize Solana block monitoring to use websocket subscriptions instead of polling.
- [ ] Secure monero view-key generation via HSM/KMS keys.
