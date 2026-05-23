# Engineering Specification: SettlerEngine Architecture
This document defines the mandatory, opinionated architectural blueprint for **SettlerEngine**—an open-source, AGPLv3-compliant, self-hosted cryptocurrency payment gateway written in pure Go.
## 1. Core Architectural Pillars
All development agents must strictly adhere to the following three foundational constraints:
 * **Sovereign Single-Binary Execution:** The application must compile down to a single, zero-dependency native binary. By default, it must require zero external operational dependencies (no external database containers, no external key-value stores).
 * **Hexagonal Architecture (Ports and Adapters):** Core domain and business logic must be isolated from databases, network protocols, transport layers, and blockchain-specific daemons.
 * **Pure Go (CGO-Free Compilation):** Under no circumstances may CGO be introduced. The code must cross-compile seamlessly using standard environment variables (e.g., GOOS=linux GOARCH=arm64 go build).
## 2. Supported Cryptographic Networks & Protocols
The core engine must abstract and monitor transactions across six distinct assets, split into three specific architectural execution engines:
```
┌──────────────────────────────────────────────────────────────────────┐
│                       SETTLERENGINE AUTOMATION                       │
└──────────────────────────────────────────────────────────────────────┘
   │
   ├── [UTXO Core Engine] ────────► BTC & LTC  (Bitcoin Core / LND / RPC)
   │
   ├── [Privacy Core Engine] ─────► XMR  (Monero Wallet RPC / View Keys)
   │
   └── [Token Accounts Engine] ───► SOL, USDC, USDT (Solana SPL / Tron TRC-20)

```
### Protocol Compliance Matrix
 1. **Bitcoin (BTC):** Native on-chain address tracking via RPC + Lightning Network invoice generation/settlement via LND gRPC streams.
 2. **Litecoin (LTC):** Native on-chain address tracking using a matching UTXO execution pipeline via litecoind RPC.
 3. **Monero (XMR):** Non-custodial invoice generation and scanning using the merchant’s **Private View Key** and public address via monero-wallet-rpc. Spend keys must never be collected or stored.
 4. **Solana (SOL) & Stablecoins (USDC/USDT):** Native Solana account tracking and SPL token monitoring utilizing pure-Go RPC clients.
 5. **Tron (TRX) & Stablecoins (USDT):** TRC-20 token transfer monitoring using pure-Go Hex/Protobuf parsers over Tron public/private RPC nodes.
## 3. Storage Layer Architecture: Dual-Database Portability
The data storage system must prioritize single-file embedded execution for self-hosters while remaining compatible with distributed relational clusters for high-volume enterprise operations.
### Mandatory Database Driver Choices
 * **Embedded (Default):** github.com/ncruces/go-sqlite3 — A 100% pure-Go, CGO-free SQLite driver utilizing WebAssembly through the wazero runtime.
 * **Distributed (Enterprise Scaling):** github.com/jackc/pgx/v5 — A pure-Go PostgreSQL driver.
### Database Compilation & Synthesis Layout (sqlc)
Do not write runtime string concatenations or introduce heavy, magic-driven Object-Relational Mappers (ORMs). Use **sqlc** to generate type-safe Go code directly from strict, portable SQL definitions.
 * **Schema Universality Rules:** All database tables must use data schemas that satisfy both SQLite and PostgreSQL strictness profiles concurrently.
 * **Identifiers:** Universal Type-Safe IDs must be managed at the application layer using **UUIDv4** or **ULID** byte arrays or strings. Do not use native database-specific auto-increment mechanisms.
 * **Timestamps:** All time markers must be stored as UTC Unix epoch integers (BIGINT) or standardized RFC3339 strings (TEXT / TIMESTAMP WITH TIME ZONE).
 * **Concurrency Mode:** When initializing SQLite, the engine must explicitly enforce **Write-Ahead Logging (WAL)** and configure a definitive busy timeout:
   ```sql
   PRAGMA journal_mode = WAL;
   PRAGMA busy_timeout = 5000;
   PRAGMA auto_vacuum = INCREMENTAL;
   
   
   ```
```

---

## 4. Automatic Pruning Mechanics

To prevent disk bloating on low-overhead infrastructure (such as standard KVM instances), the storage driver adapter must encapsulate a decoupled automatic maintenance pipeline.


```
┌────────────────────────────────────────────────────────┐
│             BACKGROUND PRUNING TIMELINE                │
└────────────────────────────────────────────────────────┘
[Every 24 Hours]
│
├──► Purge unpaid invoices > 48 hours old
│
├──► Execute SQLite incremental page vacuum
│
└──► Archive settled ledgers > 365 days old to flat files
```

The database storage adapter must run a low-priority background ticker executing the following procedures:
1.  **Expired Row Purging:** Hard delete all invoices matching status = 'expired' where the creation timestamp exceeds 48 hours.
2.  **Incremental OS Page Release:** Execute PRAGMA incremental_vacuum(100); immediately following a purge event to release empty pages back to the host operating system without initiating table-wide locking sequences.
3.  **Flat-File Ledger Archiving:** Settled transactional logs exceeding 365 days must be extracted, compiled into a compressed flat-file append-only log (archive_[year].json), and dropped from the operational live database to maintain slim memory-mapped page lookups.

---

## 5. Hexagonal Directory Structural Protocol

The project directory structure must systematically isolate core business domain concepts from infrastructure targets using strict decoupled packages.


```
settlerengine/
├── cmd/
│   └── settlerd/            # Application entry point (Main loop, initialization)
├── internal/
│   ├── domain/              # Pure business entities & abstract core types
│   │   ├── invoice.go       # Structs representing Invoice, Payment, Status
│   │   └── currency.go      # Currency-specific domain models
│   ├── ports/               # Strict behavioral definitions (Interfaces)
│   │   ├── database.go      # DB operations (Read/Write definitions)
│   │   └── blockchain.go    # Node interactions, address generation, tracking
│   ├── adapters/            # Concreted implementations of external boundary lines
│   │   ├── storage/         # Database-specific engine logic
│   │   │   ├── sqlite/      # ncruces/go-sqlite3 logic generated via sqlc
│   │   │   └── postgres/    # jackc/pgx implementation generated via sqlc
│   │   ├── crypto/          # Protocol parsers & client bindings
│   │   │   ├── bitcoin/     # BTC RPC & LND gRPC clients
│   │   │   ├── monero/      # Private view-key scanners
│   │   │   └── solana/      # SPL Token trackers
│   │   └── http/            # Web servers, merchant dashboard, webhooks
│   └── service/             # Orchestration layers binding ports to domain flows
│       └── payment_engine.go# Process flows (e.g., CreateInvoice, HandlePayment)
└── sql/
├── schema.sql           # Shared, portable table blueprints
└── queries.sql          # SQL statements parsed by sqlc
```

### Interface Inversion Enforcements

*   **Boundary Rule:** Files residing within internal/domain or internal/ports must never import external infrastructure components, drivers, or packages located inside internal/adapters.
*   **Database Inversion Factory Pattern:** The core system initialization sequence must bind dependencies dynamically based on incoming environment configurations:

```go
package main

import (
    "context"
    "log"
    "os"

    "settlerengine/internal/adapters/storage/postgres"
    "settlerengine/internal/adapters/storage/sqlite"
    "settlerengine/internal/ports"
)

func InitializeStore(ctx context.Context) ports.InvoiceStore {
    dbType := os.Getenv("SETTLER_DB_TYPE")

    switch dbType {
    case "postgres":
        return postgres.NewRepository(os.Getenv("DATABASE_URL"))
    case "sqlite", "":
        return sqlite.NewRepository(os.Getenv("SQLITE_DB_PATH"))
    default:
        log.Fatalf("Invalid database implementation selection: %s", dbType)
        return nil
    }
}

```
## 6. Real-Time Execution Boundaries & Limits
When programming, test and allocate resource strategies to guarantee execution safety up to these theoretical runtime baselines:
 * **Concurrency Handling Performance:** SQLite configurations must safely queue up to 5 seconds of active blocking delays using the WAL-backed driver. Business layer services must assume a writing limit threshold of **150–300 consecutive database mutations per second**.
 * **Resource Footprint Ceilings:** The core web server, internal memory footprint, caching tables, and data tracking routers must consume less than **100MB of operational RAM** when running under peak invoice generation load.
 * **High-Availability Scaling Configuration:** If horizontal architecture scaling across multiple systems is required, do not refactor the inner core code. Instruct infrastructure layers to deploy the identical vanilla compiled binary inside a **LiteFS** FUSE wrapper layer, routing write processes to a single designated primary controller, or transition configuration directly to a PostgreSQL destination cluster.
