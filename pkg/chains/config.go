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
		Name:        "Base",
		ChainID:     ChainIDBase,
		RPCURL:      "https://mainnet.base.org",
		USDCAddress: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
		ExplorerURL: "https://basescan.org",
	})
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
	RegisterChain(ChainConfig{
		Name:        "Avalanche",
		ChainID:     ChainIDAvalanche,
		RPCURL:      "https://api.avax.network/ext/bc/C/rpc",
		USDCAddress: "0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E",
		ExplorerURL: "https://snowtrace.io",
	})
	RegisterChain(ChainConfig{
		Name:        "Polygon",
		ChainID:     ChainIDPolygon,
		RPCURL:      "https://polygon-rpc.com",
		USDCAddress: "0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359", // Native USDC
		ExplorerURL: "https://polygonscan.com",
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
