package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/pbkdf2"
)

type EncryptedKeystore struct {
	Ciphertext []byte `json:"ciphertext"`
	Nonce      []byte `json:"nonce"`
	Salt       []byte `json:"salt"`
}

// GenerateSessionKey creates a new random session key and saves it encrypted into a keystore file.
func GenerateSessionKey(filePath string, password string) (string, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate ecdsa key: %w", err)
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Derive AES-256 key from password using PBKDF2 with SHA-256
	aesKey := pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nil, nonce, privateKeyBytes, nil)

	keystore := EncryptedKeystore{
		Ciphertext: ciphertext,
		Nonce:      nonce,
		Salt:       salt,
	}

	data, err := json.Marshal(keystore)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return "", fmt.Errorf("failed to save keystore file: %w", err)
	}

	return address, nil
}

// LoadSessionKey decrypts a keystore file back to an ECDSA private key.
func LoadSessionKey(filePath string, password string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore file: %w", err)
	}

	var keystore EncryptedKeystore
	if err := json.Unmarshal(data, &keystore); err != nil {
		return nil, err
	}

	// Re-derive the AES-256 key
	aesKey := pbkdf2.Key([]byte(password), keystore.Salt, 4096, 32, sha256.New)

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	privateKeyBytes, err := aesGCM.Open(nil, keystore.Nonce, keystore.Ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt keystore (invalid password?): %w", err)
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid private key data parsed: %w", err)
	}

	return privateKey, nil
}
