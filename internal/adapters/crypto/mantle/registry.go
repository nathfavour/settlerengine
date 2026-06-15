package mantle

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"strings"
)

// SettlerRegistryABI is the ABI for the logAgentPayment method.
const SettlerRegistryABI = `[{"inputs":[{"internalType":"bytes32","name":"_agentId","type":"bytes32"},{"internalType":"uint256","name":"_invoiceId","type":"uint256"},{"internalType":"uint256","name":"_amount","type":"uint256"},{"internalType":"string","name":"_metadata","type":"string"}],"name":"logAgentPayment","outputs":[],"stateMutability":"external","type":"function"}]`

type RegistryClient struct {
	client   *ethclient.Client
	contract common.Address
	chainID  *big.Int
}

func NewRegistryClient(rpcURL string, contractAddr common.Address, chainID *big.Int) (*RegistryClient, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}
	return &RegistryClient{
		client:   client,
		contract: contractAddr,
		chainID:  chainID,
	}, nil
}

func (c *RegistryClient) LogPayment(ctx context.Context, hexKey string, agentID [32]byte, invoiceID *big.Int, amount *big.Int, metadata string) (string, error) {
	privateKey, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return "", err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := c.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return "", err
	}

	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return "", err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, c.chainID)
	if err != nil {
		return "", err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(300000)
	auth.GasPrice = gasPrice

	parsedABI, err := abi.JSON(strings.NewReader(SettlerRegistryABI))
	if err != nil {
		return "", err
	}

	contract := bind.NewBoundContract(c.contract, parsedABI, c.client, c.client, c.client)
	tx, err := contract.Transact(auth, "logAgentPayment", agentID, invoiceID, amount, metadata)
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}
