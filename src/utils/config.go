package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	d8x_config "github.com/D8-X/d8x-futures-go-sdk/config"
	"github.com/ethereum/go-ethereum/common"
)

// load configuration json with deployment addresses: "config/chainConfig.json"
func LoadChainConfig(configName string) (map[int64]ChainConfig, error) {
	// Read the JSON file
	data, err := os.ReadFile(configName)
	if err != nil {
		log.Fatal("Error reading JSON file:", err)
		return nil, err
	}
	var configuration []ChainConfigFile
	// Unmarshal the JSON data into the Configuration struct
	err = json.Unmarshal(data, &configuration)
	if err != nil {
		log.Fatal("Error decoding JSON:", err)
		return nil, err
	}
	config := make(map[int64]ChainConfig)
	for k := range configuration {
		if len(configuration[k].AllowedExecutors) == 0 {
			// we have no executors whitelisted
			continue
		}
		sdkConf, err := d8x_config.GetDefaultChainConfigFromId(configuration[k].ChainId)
		if err != nil {
			return nil, fmt.Errorf("unable to find sdk chain config for chain %d", configuration[k].ChainId)
		}
		if sdkConf.MultiPayAddr == (common.Address{}) {
			return nil, fmt.Errorf("no multipay addr defined in sdk chain config for chain %d", configuration[k].ChainId)
		}
		if sdkConf.ProxyAddr == (common.Address{}) {
			return nil, fmt.Errorf("no proxy defined in sdk chain config for chain %d", configuration[k].ChainId)
		}
		config[configuration[k].ChainId] = ChainConfig{
			ChainId:           configuration[k].ChainId,
			Name:              configuration[k].Name,
			AllowedExecutors:  configuration[k].AllowedExecutors,
			MultiPayCtrctAddr: sdkConf.MultiPayAddr,
			//ProxyAddr:         sdkConf.ProxyAddr,
		}
	}
	return config, nil
}

// load configuration json with deployment addresses: "config/rpcConfig.json"
func LoadRpcConfig(configName string) ([]RpcConfig, error) {
	// Read the JSON file
	data, err := os.ReadFile(configName)
	if err != nil {
		log.Fatal("Error reading JSON file:", err)
		return []RpcConfig{}, err
	}
	var configuration []RpcConfig
	// Unmarshal the JSON data into the Configuration struct
	err = json.Unmarshal(data, &configuration)
	if err != nil {
		log.Fatal("Error decoding JSON:", err)
		return []RpcConfig{}, err
	}
	return configuration, nil
}
