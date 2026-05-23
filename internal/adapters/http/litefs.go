package http

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// LiteFSWriteRouting checks if the database is running under a LiteFS mount
// and intercepts write requests (POST, PUT, DELETE, PATCH) on read replicas.
// It automatically forwards them using Fly-Replay headers or redirect status.
func LiteFSWriteRouting(primaryFilePath string, next http.Handler) http.Handler {
	if primaryFilePath == "" {
		primaryFilePath = "/litefs/.primary"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Safe methods (reads) are served locally by the replica
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		// 2. Read the LiteFS .primary lease file to identify the active primary node
		primaryNode, isReplica, err := checkIsReplica(primaryFilePath)
		if err != nil {
			log.Printf("⚠️ LiteFS: Error checking primary status: %v. Defaulting to local write.", err)
			next.ServeHTTP(w, r)
			return
		}

		// 3. If this instance is a replica, route the write request to the primary node
		if isReplica {
			log.Printf("🔄 LiteFS: Routing write request (%s %s) from replica to primary node: %s", r.Method, r.URL.Path, primaryNode)

			// If running on Fly.io, we can trigger an automatic platform replay on the primary instance
			if os.Getenv("FLY_APP_NAME") != "" {
				w.Header().Set("fly-replay", "instance=primary")
				w.WriteHeader(http.StatusAccepted)
				return
			}

			// Fallback: If not on Fly.io, return a 307 Temporary Redirect to the primary node's address
			// Clients will replay the exact request (method, body, headers) on the primary node.
			primaryURL := fmt.Sprintf("http://%s:8080%s", primaryNode, r.URL.RequestURI())
			w.Header().Set("Location", primaryURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
			_, _ = io.WriteString(w, "Redirecting write request to LiteFS primary node.")
			return
		}

		// 4. We are the primary node; proceed with writing locally
		next.ServeHTTP(w, r)
	})
}

// checkIsReplica reads the LiteFS lease file and compares it against the local hostname
// to determine if the current instance is a read-only replica.
func checkIsReplica(path string) (string, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No LiteFS mount detected; running in standalone mode
			return "", false, nil
		}
		return "", false, err
	}

	primaryHost := strings.TrimSpace(string(data))
	if primaryHost == "" {
		// Empty primary file implies this node currently holds the lease (is the primary)
		return "", false, nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		return primaryHost, true, nil
	}

	// If the lease primary hostname differs from our hostname, we are a replica
	if primaryHost != hostname {
		return primaryHost, true, nil
	}

	return "", false, nil
}
