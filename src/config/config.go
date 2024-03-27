package config

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/D8-X/d8x-broker-server/src/utils"
	d8x_config "github.com/D8-X/d8x-futures-go-sdk/config"
	"github.com/ethereum/go-ethereum/common"
)

// load configuration json with deployment addresses: "config/chainConfig.json"
func LoadChainConfig(configName string) ([]utils.ChainConfig, error) {
	// Read the JSON file
	data, err := os.ReadFile(configName)
	if err != nil {
		log.Fatal("Error reading JSON file:", err)
		return []utils.ChainConfig{}, err
	}
	var configuration []utils.ChainConfigFile
	// Unmarshal the JSON data into the Configuration struct
	err = json.Unmarshal(data, &configuration)
	if err != nil {
		log.Fatal("Error decoding JSON:", err)
		return []utils.ChainConfig{}, err
	}
	config := make([]utils.ChainConfig, len(configuration))
	for k := 0; k < len(configuration); k++ {
		mAddr, err := d8x_config.GetMultiPayAddr(configuration[k].ChainId)
		var mAddrC common.Address
		if err != nil {
			msg := fmt.Sprintf("could not find multipayaddr from sdk for chain %d", configuration[k].ChainId)
			slog.Info(msg)
			mAddrC = common.Address{}
		} else {
			mAddrC = common.HexToAddress(mAddr)
		}
		config[k] = utils.ChainConfig{
			ChainId:           configuration[k].ChainId,
			Name:              configuration[k].Name,
			AllowedExecutors:  configuration[k].AllowedExecutors,
			MultiPayCtrctAddr: mAddrC,
		}
	}
	return config, nil
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
