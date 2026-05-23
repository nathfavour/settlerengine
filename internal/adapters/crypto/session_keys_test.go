package crypto

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

func TestSessionKeyKeystorePipeline(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "session.key")
	password := "correcthorsebatterystaple"

	// 1. Generate encrypted session key
	address, err := GenerateSessionKey(filePath, password)
	if err != nil {
		t.Fatalf("failed to generate session key: %v", err)
	}

	if address == "" {
		t.Errorf("expected generated address to not be empty")
	}

	// 2. Load decrypted key back
	loadedKey, err := LoadSessionKey(filePath, password)
	if err != nil {
		t.Fatalf("failed to load session key: %v", err)
	}

	loadedAddress := crypto.PubkeyToAddress(loadedKey.PublicKey).Hex()
	if loadedAddress != address {
		t.Errorf("address mismatch: generated %s, loaded %s", address, loadedAddress)
	}

	// 3. Test failure on invalid password decryption
	_, err = LoadSessionKey(filePath, "wrongpassword")
	if err == nil {
		t.Errorf("expected error when decrypting with wrong password, but got none")
	}
}
