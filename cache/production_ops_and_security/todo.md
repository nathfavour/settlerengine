# Todo Checklist: Production Ops & Security

## Short-Term Goals (1-2 Weeks)
- [x] Configure LiteFS test deployment configuration files.
- [x] Implement public key derivation (XPUB/ZPUB) address generators in Go.
- [ ] Create Prometheus instrumentation to monitor operational memory footprints.

## Medium-Term Goals (1 Month)
- [ ] Build standard PSBT export/import workflows in `settlerd` for multi-sig.
- [ ] Secure Warm Wallet using vault/secret-manager integrations.
- [x] Implement write-queue rate-limiters on SQLite to guarantee WAL locks do not exceed 5 seconds.

## Long-Term Goals (Production Gating)
- [ ] Deploy secondary read replicas behind LiteFS replication stream.
- [ ] Conduct rigorous fuzz and performance load testing to ensure RAM remains <100MB at 300 requests/sec.
- [ ] Setup zero-knowledge / MPC thresholds for warm wallet automation keys.
