# Todo Checklist: Agent Economy

## Short-Term Goals (1-2 Weeks)
- [ ] Define dynamic pricing rules in the Go server configuration.
- [ ] Scaffold the CLI structure of the `agent-sidecar` tool in Go.
- [ ] Draft specifications for SLA escrow-lock contract state logic.

## Medium-Term Goals (1 Month)
- [ ] Complete the `agent-sidecar` local wallet integration (securing private key storage, automatic EIP-712 signing).
- [ ] Integrate request-gating middleware with dynamic pricing rules.
- [ ] Support custom reputation matrices for dynamic client pricing.

## Long-Term Goals (Production Gating)
- [ ] Build the smart contract escrow layer on Base and Solana to support optimistic SLA settlement.
- [ ] Implement zero-knowledge proofs (ZKP) or optimistic verification signatures for delivery gating.
- [ ] Secure sidecar credentials using macOS Keychain, Linux Keyring, or Windows Credential Manager.
