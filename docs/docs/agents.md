# Agent Integration Guide

AI Agents can integrate with SettlerEngine to perform autonomous, cryptographically secure payments for digital resources.

## 1. The HTTP Handshake (x402)

When an agent hits a protected endpoint, it receives an `HTTP 402 Payment Required` response.

### Step A: Receive Challenge
```json
{
  "status": 402,
  "title": "Payment Required",
  "description": "This resource requires a valid x402 payment signature.",
  "payment": {
    "amount": "1000000",
    "asset": "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
    "network": "84532",
    "recipient": "0x1234567890AbcdEF1234567890aBcdef12345678",
    "nonce": "c4e9de00cbdd804fdc7fd131701a2975"
  }
}
```

### Step B: Sign Intent to Pay
The agent must sign an EIP-712 message containing the following fields:
- `recipient`: The merchant wallet address.
- `amount`: Atomic units of the asset.
- `asset`: Contract address of the token (e.g., USDC).
- `nonce`: The unique session identifier provided in the challenge.
- `deadline`: A Unix timestamp after which the signature is invalid.

### Step C: Retry with X-Payment
The agent sends the original request again, including the `X-Payment` header with the JSON-encoded `intent` and `signature`.

```http
GET /protected-resource HTTP/1.1
X-Payment: {"intent": {...}, "signature": "0x..."}
```

## 2. Local Integration (Unix Domain Sockets)

For agents running on the same host as SettlerEngine, the Unix Domain Socket (UDS) provides a low-latency, secure channel.

### Socket Location
- **Linux**: `~/.config/settlerengine/settler.sock`
- **macOS**: `~/Library/Application Support/settlerengine/settler.sock`

### Benefits of UDS
- **Zero Network Overhead**: Faster communication for high-frequency agents.
- **Local Trust**: No need for complex network configurations or TLS for local traffic.
- **Persistence**: Verified states are automatically cached in the local SQLite database.

## 3. Best Practices for Agents
- **Nonce Management**: Always use the most recent nonce provided in the 402 challenge.
- **Deadline Handling**: Set a reasonable deadline (e.g., +5 minutes) to avoid signature expiration during processing.
- **Error Handling**: Be prepared to handle signature verification failures by requesting a new nonce.
