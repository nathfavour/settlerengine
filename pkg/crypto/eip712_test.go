package crypto

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

func TestVerifyIntentToPay(t *testing.T) {
	// 1. Setup private key and address
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)

	// 2. Define Domain and Intent
	params := DomainParams{
		ChainID:           big.NewInt(8453), // Base
		VerifyingContract: common.HexToAddress("0x1234567890123456789012345678901234567890"),
	}

	intent := IntentToPay{
		Recipient: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B",
		Amount:    "1000000", // 1 USDC (6 decimals)
		Asset:     "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", // USDC
		Nonce:     "test-nonce-123",
		Deadline:  1739686400,
	}

	// 3. Manually create the hash for signing (mirroring logic in eip712.go)
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"IntentToPay": []apitypes.Type{
				{Name: "recipient", Type: "address"},
				{Name: "amount", Type: "uint256"},
				{Name: "asset", Type: "address"},
				{Name: "nonce", Type: "string"},
				{Name: "deadline", Type: "int64"},
			},
		},
		PrimaryType: "IntentToPay",
		Domain: apitypes.TypedDataDomain{
			Name:              "SettlerEngine",
			Version:           "1",
			ChainId:           (*math.HexOrDecimal256)(params.ChainID),
			VerifyingContract: params.VerifyingContract.Hex(),
		},
		Message: apitypes.TypedDataMessage{
			"recipient": intent.Recipient,
			"amount":    intent.Amount,
			"asset":     intent.Asset,
			"nonce":     intent.Nonce,
			"deadline":  intent.Deadline,
		},
	}

	domainSeparator, _ := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	typedDataHash, _ := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	rawData := []byte("\x19\x01" + string(domainSeparator) + string(typedDataHash))
	sighash := crypto.Keccak256(rawData)

	// 4. Sign the hash
	signatureBytes, err := crypto.Sign(sighash, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	// Adjust V to be 27/28
	signatureBytes[64] += 27
	signature := hexutil.Encode(signatureBytes)

	// 5. Verify
	recoveredAddr, err := VerifyIntentToPay(intent, signature, params)
	if err != nil {
		t.Fatalf("Verification failed: %v", err)
	}

	if recoveredAddr != addr {
		t.Errorf("Expected address %s, got %s", addr.Hex(), recoveredAddr.Hex())
	}
}
