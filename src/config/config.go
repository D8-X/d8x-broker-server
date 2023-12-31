package config

import (
	"encoding/json"
	"log"
	"os"

	"github.com/D8-X/d8x-broker-server/src/utils"
)

// load configuration json with deployment addresses: "config/chainConfig.json"
func LoadChainConfig(configName string) ([]utils.ChainConfig, error) {
	// Read the JSON file
	data, err := os.ReadFile(configName)
	if err != nil {
		log.Fatal("Error reading JSON file:", err)
		return []utils.ChainConfig{}, err
	}
	var configuration []utils.ChainConfig
	// Unmarshal the JSON data into the Configuration struct
	err = json.Unmarshal(data, &configuration)
	if err != nil {
		log.Fatal("Error decoding JSON:", err)
		return []utils.ChainConfig{}, err
	}
	return configuration, nil
}

// load configuration json with deployment addresses: "config/rpcConfig.json"
func LoadRpcConfig(configName string) ([]utils.RpcConfig, error) {
	// Read the JSON file
	data, err := os.ReadFile(configName)
	if err != nil {
		log.Fatal("Error reading JSON file:", err)
		return []utils.RpcConfig{}, err
	}
	var configuration []utils.RpcConfig
	// Unmarshal the JSON data into the Configuration struct
	err = json.Unmarshal(data, &configuration)
	if err != nil {
		log.Fatal("Error decoding JSON:", err)
		return []utils.RpcConfig{}, err
	}
	return configuration, nil
}
