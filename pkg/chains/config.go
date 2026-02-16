package chains

import (
	"fmt"
)

type ChainID uint64

const (
	ChainIDEthereum      ChainID = 1
	ChainIDBase          ChainID = 8453
	ChainIDCronos        ChainID = 25
	ChainIDAvalanche     ChainID = 43114
	ChainIDPolygon       ChainID = 137
	ChainIDBaseSepolia   ChainID = 84532
	ChainIDCronoszkEVM   ChainID = 240
)

type ChainConfig struct {
	Name               string
	ChainID            ChainID
	RPCURL             string
	FacilitatorAddress string
	USDCAddress        string
	ExplorerURL        string
}

func init() {
	RegisterChain(ChainConfig{
		Name:        "Base Sepolia",
		ChainID:     ChainIDBaseSepolia,
		RPCURL:      "https://sepolia.base.org",
		USDCAddress: "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		ExplorerURL: "https://sepolia.basescan.org",
	})
	RegisterChain(ChainConfig{
		Name:        "Cronos zkEVM Testnet",
		ChainID:     ChainIDCronoszkEVM,
		RPCURL:      "https://cronos-zkevm-testnet.drpc.org",
		USDCAddress: "0xaa5b845F8C9c047779bEDf64829601d8B264076c",
		ExplorerURL: "https://explorer.zkevm.cronos.org/testnet/",
	})
}

var registry = make(map[ChainID]ChainConfig)

func RegisterChain(cfg ChainConfig) {
	registry[cfg.ChainID] = cfg
}

func GetChainConfig(id ChainID) (ChainConfig, error) {
	cfg, ok := registry[id]
	if !ok {
		return ChainConfig{}, fmt.Errorf("chain ID %d not supported", id)
	}
	return cfg, nil
}
