package chains

import (
	"fmt"
)

type ChainID uint64

const (
	ChainIDEthereum ChainID = 1
	ChainIDBase     ChainID = 8453
	ChainIDCronos   ChainID = 25
	ChainIDAvalanche ChainID = 43114
	ChainIDPolygon  ChainID = 137
)

type ChainConfig struct {
	Name               string
	ChainID            ChainID
	RPCURL             string
	FacilitatorAddress string
	USDCAddress        string
	ExplorerURL        string
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
