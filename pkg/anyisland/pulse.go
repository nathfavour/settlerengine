package anyisland

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

type HandshakeOp struct {
	Op string `json:"op"`
}

type HandshakeResponse struct {
	Status           string `json:"status"`
	ToolID           string `json:"tool_id,omitempty"`
	Version          string `json:"version,omitempty"`
	AnyislandVersion string `json:"anyisland_version,omitempty"`
}

type RegisterOp struct {
	Op      string `json:"op"`
	Name    string `json:"name"`
	Source  string `json:"source"`
	Version string `json:"version"`
	Type    string `json:"type"`
}

// CheckManaged connects to the Anyisland socket to verify if the tool is managed.
func CheckManaged() (*HandshakeResponse, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	sockPath := filepath.Join(home, ".anyisland", "anyisland.sock")
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		return nil, fmt.Errorf("could not connect to anyisland socket: %w", err)
	}
	defer conn.Close()

	req := HandshakeOp{Op: "HANDSHAKE"}
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return nil, err
	}

	var resp HandshakeResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// Register sends a registration packet to the local Anyisland daemon.
func Register(name, version string) error {
	packet := RegisterOp{
		Op:      "REGISTER",
		Name:    name,
		Source:  "github.com/nathfavour/settlerengine",
		Version: version,
		Type:    "binary",
	}
	data, err := json.Marshal(packet)
	if err != nil {
		return err
	}

	conn, err := net.DialTimeout("udp", "localhost:1995", 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(data)
	return err
}
