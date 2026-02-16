package crypto

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// IntentToPay represents the x402 handshake payload signed by an agent.
type IntentToPay struct {
	Recipient string `json:"recipient"` // Merchant wallet address
	Amount    string `json:"amount"`    // Amount in atomic units (uint256 string)
	Asset     string `json:"asset"`     // Token contract address (USDC)
	Nonce     string `json:"nonce"`     // Unique session UUID to prevent replay
	Deadline  uint64 `json:"deadline"`  // Unix timestamp for signature expiry
}

// DomainParams defines the parameters for the EIP-712 domain separator.
type DomainParams struct {
	ChainID           *big.Int
	VerifyingContract common.Address
}

// VerifyIntentToPay checks if the signature is valid for the given intent and domain.
func VerifyIntentToPay(intent IntentToPay, signature string, params DomainParams) (common.Address, error) {
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
				{Name: "deadline", Type: "uint256"},
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
			"deadline":  (*math.HexOrDecimal256)(big.NewInt(int64(intent.Deadline))),
		},
	}

	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to hash domain separator: %w", err)
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to hash message: %w", err)
	}

	rawData := make([]byte, 2+32+32)
	rawData[0] = 0x19
	rawData[1] = 0x01
	copy(rawData[2:34], domainSeparator)
	copy(rawData[34:66], typedDataHash)
	sighash := crypto.Keccak256(rawData)

	sig, err := hexutil.Decode(signature)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to decode signature: %w", err)
	}

	if len(sig) != 65 {
		return common.Address{}, fmt.Errorf("invalid signature length: %d", len(sig))
	}

	if sig[64] >= 27 {
		sig[64] -= 27
	}

	pubKey, err := crypto.SigToPub(sighash, sig)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to recover public key: %w", err)
	}

	return crypto.PubkeyToAddress(*pubKey), nil
}
