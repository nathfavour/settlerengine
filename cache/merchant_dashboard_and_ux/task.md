# Merchant Dashboard & Human Experience (UX)

To match the robustness of BTCPay Server, SettlerEngine requires an exceptional human interface that provides real-time visualization of payments, analytics, yield performances, and agent actions.

## Milestones & Goals

### 1. Headless Merchant Dashboard (Next.js/React)
*   Build a sleek modern frontend communicating with `settlerd` over REST APIs.
*   Visualize real-time payment streams with smooth animations.
*   Render metrics of Automated Yield: TVL, earnings, current BSC Vault APYs, and time-to-settle charts.
*   Include log views for x402 handshakes (gated resources, failure codes, nonces).

### 2. Enterprise Webhook Delivery System
*   Implement standard webhook publishers notifying legacy backends (e.g., standard e-commerce carts) of confirmed settlements.
*   Enforce signature verification (e.g., HMAC-SHA256) on all webhook request headers to prevent spoofing.
*   Ensure transactional reliability with retry timers (exponential backoff) and persistent queueing in the database.

### 3. Multi-Tenant Project Isolation
*   Support multiple merchant "stores" or isolation contexts under a single running instance.
*   Setup API credentials, tokens, and verification keys per store.
