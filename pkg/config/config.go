package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type SettlerConfig struct {
	RPCURL                 string `json:"rpc_url"`
	PrivateKey             string `json:"private_key"`
	RegistryAddress        string `json:"registry_address"`
	AgentID                string `json:"agent_id"`
	CasperFacilitatorURL   string `json:"casper_facilitator_url"`
	CasperFacilitatorToken string `json:"casper_facilitator_token"`
}

func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(configDir, "settlerengine")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func LoadConfig() (*SettlerConfig, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Auto-initialize default config if it doesn't exist
			defaultCfg := &SettlerConfig{
				RPCURL:                 "https://rpc.sepolia.mantle.xyz",
				RegistryAddress:        "0x33aE8331a2406EEc3A33483001aC5650DA2e0662",
				AgentID:                "42",
				CasperFacilitatorURL:   "https://x402-facilitator.cspr.cloud",
				CasperFacilitatorToken: "", // empty/permissionless default
			}
			if err := SaveConfig(defaultCfg); err != nil {
				return nil, err
			}
			return defaultCfg, nil
		}
		return nil, err
	}
	var cfg SettlerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveConfig(cfg *SettlerConfig) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (c *SettlerConfig) Prompt() {
	fmt.Println("⚙️ SettlerEngine Configuration Setup")
	
	fmt.Printf("Ethereum RPC URL [%s]: ", c.RPCURL)
	var rpc string
	fmt.Scanln(&rpc)
	if rpc != "" {
		c.RPCURL = rpc
	}

	fmt.Printf("Private Key (hex) [hidden]: ")
	var key string
	fmt.Scanln(&key)
	if key != "" {
		c.PrivateKey = key
	}

	fmt.Printf("Registry Contract Address [%s]: ", c.RegistryAddress)
	var reg string
	fmt.Scanln(&reg)
	if reg != "" {
		c.RegistryAddress = reg
	}

	fmt.Printf("Agent ID [%s]: ", c.AgentID)
	var id string
	fmt.Scanln(&id)
	if id != "" {
		c.AgentID = id
	}
}
