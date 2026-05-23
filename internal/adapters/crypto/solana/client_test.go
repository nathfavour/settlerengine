package solana

import (
	"math/big"
	"testing"
)

func TestHexToTronAddress(t *testing.T) {
	// Let's use a known Tron address example:
	// T9yD14Nj9j7xAB4dbGeiX9h8unkKHxuWwb
	// Wait, let's derive one and test our roundtrip or check standard conversions
	hexAddr := "850d995c6b9e7fa1cc16910c21c08ec5a46444fb"
	expectedTronAddr := "TNNsTzF68nCipQyS4hT3XfubXw96S51uVd" // prefix 0x41 + hexAddr + checksum in Base58

	res, err := HexToTronAddress(hexAddr)
	if err != nil {
		t.Fatalf("HexToTronAddress failed: %v", err)
	}

	// Verify it derives a valid Tron address format starting with "T"
	if res == "" || res[0] != 'T' {
		t.Errorf("Expected address to start with 'T', got: %s", res)
	}

	// Double check padded hex topic decoding
	paddedHex := "0x000000000000000000000000850d995c6b9e7fa1cc16910c21c08ec5a46444fb"
	res2, err := HexToTronAddress(paddedHex)
	if err != nil {
		t.Fatalf("HexToTronAddress on padded hex failed: %v", err)
	}
	if res2 != res {
		t.Errorf("Padded and unpadded hex address mismatch: %s vs %s", res2, res)
	}
}

func TestParseTRC20TransferLog(t *testing.T) {
	// Construct a standard Transfer event log
	// Topic 0: Keccak256("Transfer(address,address,uint256)")
	topic0 := "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	// Topic 1: Padded Sender Address (850d995c6b9e7fa1cc16910c21c08ec5a46444fb)
	topic1 := "0x000000000000000000000000850d995c6b9e7fa1cc16910c21c08ec5a46444fb"
	// Topic 2: Padded Recipient Address (998ce54b0dd67027ec1bd743e006a7b27f106f710)
	topic2 := "0x000000000000000000000000998ce54b0dd67027ec1bd743e006a7b27f106f710"
	// Data: 10 USDT (10,000,000 units = 0x989680)
	data := "0x0000000000000000000000000000000000000000000000000000000000989680"

	logEntry := TRC20TransferLog{
		Address: "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		Topics:  []string{topic0, topic1, topic2},
		Data:    data,
	}

	from, to, val, err := ParseTRC20TransferLog(logEntry)
	if err != nil {
		t.Fatalf("ParseTRC20TransferLog failed: %v", err)
	}

	if from == "" || from[0] != 'T' {
		t.Errorf("Expected valid from address, got: %s", from)
	}
	if to == "" || to[0] != 'T' {
		t.Errorf("Expected valid to address, got: %s", to)
	}

	expectedAmount := big.NewInt(10000000)
	if val.Cmp(expectedAmount) != 0 {
		t.Errorf("Expected transfer value of %s, got: %s", expectedAmount.String(), val.String())
	}
}

func TestParseTRC20TransferLog_InvalidSignature(t *testing.T) {
	// Topic 0: Random non-transfer event signature
	topic0 := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	topic1 := "0x000000000000000000000000850d995c6b9e7fa1cc16910c21c08ec5a46444fb"
	topic2 := "0x000000000000000000000000998ce54b0dd67027ec1bd743e006a7b27f106f710"

	logEntry := TRC20TransferLog{
		Address: "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		Topics:  []string{topic0, topic1, topic2},
		Data:    "0x0000000000000000000000000000000000000000000000000000000000989680",
	}

	_, _, _, err := ParseTRC20TransferLog(logEntry)
	if err == nil {
		t.Errorf("Expected error when parsing log with invalid signature, got nil")
	}
}
