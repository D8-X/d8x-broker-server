package contracts

import (
	"context"
	"math/big"
	"strings"

	"github.com/D8-X/d8x-futures-go-sdk/utils"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var erc20ABI = `[{"constant":true,"inputs":[{"name":"owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"type":"function"}]`

func BalanceOf(client *ethclient.Client, decimals int, token, user common.Address) (float64, error) {

	// Prepare the ABI
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return 0, err
	}

	// Prepare data for the balanceOf call
	data, err := parsedABI.Pack("balanceOf", user)
	if err != nil {
		return 0, err
	}

	// Call contract
	msg := ethereum.CallMsg{
		To:   &token,
		Data: data,
	}

	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return 0, err
	}

	// Parse result
	var balance = new(big.Int)
	err = parsedABI.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		return 0, err
	}
	b := utils.DecNToFloat(balance, uint8(decimals))
	return b, nil
}
