# Supported Chains

SettlerEngine natively supports several chains to ensure compliance with the x402 standard and enable seamless agentic commerce.

## Primary Native Support (EVM + EIP-3009)

We prioritize chains that support **EIP-3009** (`transferWithAuthorization`), allowing agents to perform gasless USDC transfers.

| Chain | Chain ID | Status | Note |
|---|---|---|---|
| **Base** | 8453 | Native | Flagship network for x402. |
| **Ethereum** | 1 | Native | High-value settlements. |
| **Cronos** | 25 | Native | Required for Cronos x Crypto.com track. |
| **Avalanche** | 43114 | Native | Optimized for sub-second finality. |
| **Polygon** | 137 | Native | Low-fee micro-settlements. |

## Emerging & Non-EVM Support

- **Solana:** Native integration via CDP and x402-solana libraries.
- **Sui & Near:** High-throughput support on the roadmap.
- **TRON & BNB Chain:** Supported for on-chain identity via ERC-8004.

## Engineering Directive

For all EVM chains, SettlerEngine uses **USDC** (implementing EIP-3009) to bypass the allowance bottleneck, ensuring a frictionless payment experience for AI agents.

## Testnet Configuration

For integration testing, use the following network parameters:

### Base Sepolia
- **Chain ID:** 84532
- **RPC:** `https://sepolia.base.org`
- **USDC Address:** `0x036CbD53842c5426634e7929541eC2318f3dCF7e`

### Cronos zkEVM (Testnet)
- **Chain ID:** 240
- **RPC:** `https://cronos-zkevm-testnet.drpc.org`
- **USDC Address:** `0xaa5b845F8C9c047779bEDf64829601d8B264076c`
