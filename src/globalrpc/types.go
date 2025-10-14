package globalrpc

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type RpcConfig struct {
	ChainId int      `json:"chainId"`
	Wss     []string `json:"WSS"`
	Https   []string `json:"HTTP"`
}

type RPCType int

const (
	TypeHTTPS RPCType = iota // Starts at 0
	TypeWSS                  // 1
)

func (t RPCType) String() string {
	switch t {
	case TypeHTTPS:
		return "HTTPS"
	case TypeWSS:
		return "WSS"
	default:
		return "Unknown"
	}
}

func loadRPCConfig(chainId int, filename string) (RpcConfig, error) {
	var config []RpcConfig
	jsonFile, err := os.Open(filename)
	if err != nil {
		return RpcConfig{}, err
	}
	defer jsonFile.Close()

	// Read the file's contents into a byte slice
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	// Unmarshal the JSON data
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return RpcConfig{}, fmt.Errorf("error unmarshalling JSON: %v", err)
	}
	idx := -1
	for j := range config {
		if config[j].ChainId == chainId {
			idx = j
			break
		}
	}
	if idx == -1 {
		return RpcConfig{}, fmt.Errorf("no rpc config for chain %d", chainId)
	}

	return config[idx], nil
}
