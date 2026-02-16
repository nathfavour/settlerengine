package x402

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	crypto2 "github.com/nathfavour/settlerengine/pkg/crypto"
)

func TestMiddleware_402Challenge(t *testing.T) {
	cfg := Config{
		DomainParams: crypto2.DomainParams{
			ChainID:           big.NewInt(84532),
			VerifyingContract: common.HexToAddress("0x0"),
		},
		NonceExpiry: 1 * time.Minute,
		Recipient:   "0x123",
		Asset:       "0x456",
		Amount:      "100",
	}
	mw := NewMiddleware(cfg)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := mw.Handler(nextHandler)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusPaymentRequired {
		t.Errorf("expected status 402, got %d", rr.Code)
	}

	var resp ChallengeResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Payment.Amount != "100" {
		t.Errorf("expected amount 100, got %s", resp.Payment.Amount)
	}

	if resp.Payment.Nonce == "" {
		t.Error("expected nonce in response")
	}
}

func TestMiddleware_Authorized(t *testing.T) {
	privateKey, _ := crypto.GenerateKey()
	agentAddr := crypto.PubkeyToAddress(privateKey.PublicKey)

	chainID := big.NewInt(84532)
	verifyingContract := common.HexToAddress("0x0")

	cfg := Config{
		DomainParams: crypto2.DomainParams{
			ChainID:           chainID,
			VerifyingContract: verifyingContract,
		},
		NonceExpiry: 1 * time.Minute,
		Recipient:   agentAddr.Hex(), // Just for testing
		Asset:       "0x456",
		Amount:      "100",
	}
	mw := NewMiddleware(cfg)

	// First request to get a nonce
	req1 := httptest.NewRequest("GET", "/", nil)
	rr1 := httptest.NewRecorder()
	mw.Handler(nil).ServeHTTP(rr1, req1)

	var challenge ChallengeResponse
	json.Unmarshal(rr1.Body.Bytes(), &challenge)
	nonce := challenge.Payment.Nonce

	// Sign the intent
	intent := crypto2.IntentToPay{
		Recipient: cfg.Recipient,
		Amount:    cfg.Amount,
		Asset:     cfg.Asset,
		Nonce:     nonce,
		Deadline:  uint64(time.Now().Add(1 * time.Hour).Unix()),
	}

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
			ChainId:           (*math.HexOrDecimal256)(chainID),
			VerifyingContract: verifyingContract.Hex(),
		},
		Message: apitypes.TypedDataMessage{
			"recipient": intent.Recipient,
			"amount":    intent.Amount,
			"asset":     intent.Asset,
			"nonce":     intent.Nonce,
			"deadline":  (*math.HexOrDecimal256)(big.NewInt(int64(intent.Deadline))),
		},
	}

	domainSeparator, _ := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	typedDataHash, _ := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	rawData := make([]byte, 2+32+32)
	rawData[0] = 0x19
	rawData[1] = 0x01
	copy(rawData[2:34], domainSeparator)
	copy(rawData[34:66], typedDataHash)
	sighash := crypto.Keccak256(rawData)

	signature, _ := crypto.Sign(sighash, privateKey)
	signature[64] += 27 // Transform V
	sigHex := "0x" + common.Bytes2Hex(signature)

	payload := PaymentPayload{
		Intent:    intent,
		Signature: sigHex,
	}
	payloadJSON, _ := json.Marshal(payload)

	// Second request with X-Payment header
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("X-Payment", string(payloadJSON))
	rr2 := httptest.NewRecorder()

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	mw.Handler(nextHandler).ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", rr2.Code, rr2.Body.String())
	}
	if !nextCalled {
		t.Error("next handler was not called")
	}

	// Third request (Idempotency check)
	req3 := httptest.NewRequest("GET", "/", nil)
	req3.Header.Set("X-Payment", string(payloadJSON))
	rr3 := httptest.NewRecorder()
	nextCalled = false
	mw.Handler(nextHandler).ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusOK {
		t.Errorf("expected status 200 (cached), got %d", rr3.Code)
	}
	if !nextCalled {
		t.Error("next handler was not called (cached)")
	}
}
