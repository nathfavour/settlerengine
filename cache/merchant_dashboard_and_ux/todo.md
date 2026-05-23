# Todo Checklist: Merchant Dashboard & UX

## Short-Term Goals (1-2 Weeks)
- [x] Define JSON-RPC / REST endpoints on `settlerd` for invoice retrieval and metrics.
- [x] Setup initial dashboard project skeleton with premium visual aesthetics.
- [x] Create simple webhook schema in database (`webhook_configs` and `webhook_logs`).

## Medium-Term Goals (1 Month)
- [x] Implement robust HMAC-SHA256 signing for all outgoing webhook payloads.
- [x] Build a reliable outbox worker with exponential backoff for failed webhook dispatches.
- [x] Complete the dashboard dashboard UI layout (invoice tables, yield performance chart, wallet configuration).
- [ ] Implement multi-tenant token-based authentication on backend REST endpoints.

## Long-Term Goals (Production Gating)
- [ ] Add support for WebSocket updates to stream live incoming payments directly to the dashboard.
- [ ] Integrate full project isolation (independent stores with custom RPC nodes and wallets).
- [ ] Establish automated webhook test dispatchers.
