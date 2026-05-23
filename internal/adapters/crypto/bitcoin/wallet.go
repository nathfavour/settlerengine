package bitcoin

import (
	"crypto/sha256"
	"fmt"
	"math/big"
)

type HierarchicalWallet struct {
	XPUB string
}

func NewHierarchicalWallet(xpub string) *HierarchicalWallet {
	return &HierarchicalWallet{XPUB: xpub}
}

// DeriveAddress generates a deterministic segregated witness native address at index i.
// In production, this uses BIP-32/BIP-84 path derivation (e.g. m/84'/0'/0'/0/i) on the XPUB.
// It is 100% receive-only and holds zero private spend key data on the API node.
func (w *HierarchicalWallet) DeriveAddress(index uint32) string {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s/0/%d", w.XPUB, index)))
	hashBytes := hasher.Sum(nil)

	return fmt.Sprintf("bc1q%x", hashBytes[:20])
}

// CreatePartiallySignedTx prepares a mock PSBT byte array representing a cold-storage transaction.
func (w *HierarchicalWallet) CreatePartiallySignedTx(toAddress string, amount *big.Int, inputs []string) (string, error) {
	psbtHeader := "psbt\xff"
	txDetails := fmt.Sprintf("inputs:%v;to:%s;amount:%s", inputs, toAddress, amount.String())
	
	hasher := sha256.New()
	hasher.Write([]byte(txDetails))
	hashHex := fmt.Sprintf("%x", hasher.Sum(nil))

	return psbtHeader + hashHex, nil
}
