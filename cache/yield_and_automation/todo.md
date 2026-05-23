# Todo Checklist: Yield & Automation

## Short-Term Goals (1-2 Weeks)
- [ ] Connect the `RiquidAdapter` to a mock BSC local node to test deposits/withdrawals.
- [ ] Implement the `GetSmartAccountAddress` deterministic calculation locally.
- [ ] Add CLI flags to import and encrypt/decrypt local session keys.

## Medium-Term Goals (1 Month)
- [ ] Complete ERC-4337 UserOperation validation and signing pipeline using local private key.
- [ ] Integrate with Stackup or Pimlico Paymaster RPC for gas sponsorship.
- [ ] Implement rebalancing engine logic that queries APY dynamically and swaps funds.

## Long-Term Goals (Production Gating)
- [ ] Deploy session key smart contracts for customized automated execution rules.
- [ ] Complete full end-to-end integration tests using Local Anvil BSC forks for harvesting.
- [ ] Enable dynamic slippage and path finding for high-volume cross-vault rebalancing.
