package storage

import (
	"testing"
)

func TestStorage_OpenDefault(t *testing.T) {
	db, err := OpenDefault()
	if err != nil {
		t.Fatalf("Failed to open default storage: %v", err)
	}
	defer db.Close()

	if db.DataDir == "" {
		t.Error("Expected DataDir to be set")
	}

	// Clean up if we are in a test environment
	// Note: In a real system we might want to use a temp dir for tests
}

func TestStorage_PaymentPersistence(t *testing.T) {
	db, err := OpenDefault()
	if err != nil {
		t.Fatalf("Failed to open: %v", err)
	}
	defer db.Close()

	sig := "0xabc123"
	signer := "0xsigner"
	
	err = db.RecordPayment(sig, signer, "100", "0xasset", "nonce1")
	if err != nil {
		t.Fatalf("Failed to record: %v", err)
	}

	recovered, err := db.CheckPayment(sig)
	if err != nil {
		t.Fatalf("Failed to check: %v", err)
	}

	if recovered != signer {
		t.Errorf("Expected %s, got %s", signer, recovered)
	}
}
