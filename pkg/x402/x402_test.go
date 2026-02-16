package x402_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// Constants for Base Sepolia Integration Testing
const (
	BaseSepoliaChainID = 84532
	BaseSepoliaUSDC    = "0x036CbD53842c5426634e7929541eC2318f3dCF7e"
	MockRecipient      = "0x1234567890AbcdEF1234567890aBcdef12345678" // Settler Merchant Wallet
)

func TestX402HandshakeSignature(t *testing.T) {
	// 1. Generate a Mock Agent Private Key (Simulating an AI Agent wallet)
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	agentAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	// 2. Define the IntentToPay Data
	amount := big.NewInt(1000000) // 1.00 USDC (6 decimals)
	nonce := hexutil.Encode(crypto.Keccak256([]byte("session-uuid-123")))
	deadline := int64(1735689600) // Future timestamp

	// 3. Mock the Domain Separator Hash
	// Normally calculated via typeddata.HashStruct
	domainSeparator := crypto.Keccak256([]byte("SettlerEngine-V1-BaseSepolia"))
	
	// 4. Create the Typed Data Hash (The "Challenge")
	messageHash := crypto.Keccak256(
		domainSeparator,
		common.HexToAddress(MockRecipient).Bytes(),
		amount.Bytes(),
		common.HexToAddress(BaseSepoliaUSDC).Bytes(),
		[]byte(nonce),
		big.NewInt(deadline).Bytes(),
	)

	// 5. Sign the Message (Agent side)
	signature, err := crypto.Sign(messageHash, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	// 6. Verification (SettlerEngine side)
	recoveredPubKey, err := crypto.SigToPub(messageHash, signature)
	if err != nil {
		t.Fatalf("Failed to recover public key: %v", err)
	}
	recoveredAddress := crypto.PubkeyToAddress(*recoveredPubKey)

	// 7. Assertions
	if recoveredAddress != agentAddress {
		t.Errorf("Recovered address %s does not match Agent address %s", recoveredAddress.Hex(), agentAddress.Hex())
	} else {
		fmt.Printf("✅ Handshake Verified: Agent %s authorized payment of 1.00 USDC on Base Sepolia\n", recoveredAddress.Hex())
	}
}
