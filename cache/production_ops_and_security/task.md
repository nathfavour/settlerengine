# Production Operations, Security & Scalability

A resilient payment gateway must guarantee zero downtime, resist host attacks, and preserve database consistency under high load.

## Milestones & Goals

### 1. LiteFS Replication Architecture
*   Integrate FUSE-based **LiteFS** to support horizontal scaling of SQLite database files.
*   Setup a primary-secondary setup where writes are routed to a single primary controller node while reads are distributed across secondary instances.
*   Enforce WAL (Write-Ahead Logging) consistency checks.

### 2. Tiered Keystore Security System (Hot/Warm/Cold)
*   **Hot Wallet (API Nodes):** Stored only as public receive-only keys (XPUB/ZPUB) to derive invoices. Cannot sign outgoing transactions.
*   **Warm Wallet (Signing Service):** Automates small yield harvesting operations using sharded keys or MPC.
*   **Cold Wallet (Offline):** Stores long-term reserve treasury offline. Requires manual Partially Signed Bitcoin Transactions (PSBT) or multi-sig approval.

### 3. Database Performance Tuning & SQLite WAL queueing
*   Implement pgx/v5 connection pooling and prepared statements for extreme throughput.
*   Optimize SQLite transaction queues to handle 150-300 writes per second.
*   Restrict standard operational RAM memory footprint to under 100MB under peak loads.
