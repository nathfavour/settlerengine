# Local Integration & UDS

SettlerEngine provides a robust local communication interface for co-located services and agents.

## Data Directory Structure

The engine stores its state and socket in a dedicated directory based on the OS standard (`os.UserConfigDir`):

```bash
settlerengine/
├── settler.db   # SQLite3 database (CGO-free)
└── settler.sock # Unix Domain Socket
```

## Unix Domain Socket (UDS) Protocol

The `settler.sock` allows local processes to interact with the engine. 

### Connection Example (Go)
```go
package main

import (
	"net"
	"log"
)

func main() {
	conn, err := net.Dial("unix", "/path/to/settler.sock")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	
	// Write and read protocol messages...
}
```

## Persistence via SQLite

SettlerEngine uses a CGO-free implementation of SQLite (`modernc.org/sqlite`) to ensure compatibility across all environments without needing a C toolchain.

### Verified Payments Table
The engine persists every verified signature to ensure idempotency and prevent replay attacks across restarts.

```sql
CREATE TABLE verified_payments (
    signature TEXT PRIMARY KEY,
    signer TEXT NOT NULL,
    amount TEXT NOT NULL,
    asset TEXT NOT NULL,
    nonce TEXT NOT NULL,
    verified_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Automatic Cleanup
Signatures are validated against their EIP-712 deadlines even when retrieved from the database. The engine performs periodic cleanup of expired session nonces.
