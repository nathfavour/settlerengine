package bitcoin

import (
	"math/big"
	"testing"
)

func TestHDWalletDerivationAndPSBT(t *testing.T) {
	xpub := "vpub5SLqmdv5yFZNn4u3sU48D..."
	wallet := NewHierarchicalWallet(xpub)

	// 1. Check deterministic address generation
	addr0_1 := wallet.DeriveAddress(0)
	addr0_2 := wallet.DeriveAddress(0)

	if addr0_1 != addr0_2 {
		t.Errorf("address derivation is not deterministic: got %s and %s", addr0_1, addr0_2)
	}

	// 2. Check index uniqueness
	addr1 := wallet.DeriveAddress(1)
	if addr0_1 == addr1 {
		t.Errorf("address derivation index collision: %s", addr0_1)
	}

	// 3. Generate cold-storage PSBT
	psbt, err := wallet.CreatePartiallySignedTx("bc1qrecipient", big.NewInt(1500000), []string{"input_utxo_1"})
	if err != nil {
		t.Fatalf("failed to create PSBT: %v", err)
	}

	if psbt[:5] != "psbt\xff" {
		t.Errorf("expected PSBT header to be psbt\\xff, got %s", psbt[:5])
	}
}
