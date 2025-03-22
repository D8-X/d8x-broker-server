package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/D8-X/d8x-broker-server/src/contracts"
	"github.com/D8-X/d8x-broker-server/src/globalrpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const API_URL = "https://dappapi.vip3.io/api/v1/sbt/info"
const VIP3_REDIS = "VIP"
const VIP3_INFO_EXPIRY_SEC int64 = 7 * 86_400
const FEE_EXPIRY_SEC int64 = 3600

// Reduction of broker fees for VIP3 per level (4 levels) is set in the .env-file as
// VIP3_REDUCTION_PERC="50,75,90,100"

type Vip3Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Level int `json:"level"`
	} `json:"data"`
}

// getBrokerFeeTbps returns the broker fee. traderAddr can be an empty string
// and chainId can be -1
func (a *App) getBrokerFeeTbps(traderAddr string, chainId int) uint16 {
	return a.getReducedBrokerFeeTbps(traderAddr, chainId)
}

func (a *App) getReducedBrokerFeeTbps(traderAddr string, chainId int) uint16 {
	traderAddr = strings.ToLower(traderAddr)
	if chainId == -1 || traderAddr == (common.Address{}.String()) || traderAddr == "" {
		return a.BrokerFeeTbps
	}
	f, err := RedisGetTraderFeeMultiplier(a.RedisClient.Client, chainId, traderAddr)
	var m float64
	if err != nil {
		// error/not existing, we need to get the fee multiplier onchain
		var err error
		m, err = a.fetchAndCacheFeeMultiplier(traderAddr, chainId)
		if err != nil {
			slog.Error("getReducedBrokerFeeTbps", "error", err)
		}
	} else {
		m = f.FeeMultiplier
		if time.Now().Unix()-f.TimestampSec > FEE_EXPIRY_SEC {
			// expired
			go func() {
				slog.Info("fee expired for trader", "trader", traderAddr, "chainId", chainId)
				_, err := a.fetchAndCacheFeeMultiplier(traderAddr, chainId)
				if err != nil {
					slog.Error("unable to fetch and cache fee", "error", err)
				} else {
					slog.Info("fetching fee successful")
				}
			}()
		}
	}
	return uint16(float64(a.BrokerFeeTbps) * m)
}

func (a *App) fetchAndCacheFeeMultiplier(traderAddr string, chainId int) (float64, error) {
	var err error
	m, err := a.fetchFeeReduction(traderAddr, chainId)
	if err != nil {
		return float64(1), fmt.Errorf("unable to fetch fee reduction %v", err)
	}
	err = RedisStoreTraderFeeMultiplier(a.RedisClient.Client, chainId, traderAddr, m)
	if err != nil {
		return float64(1), fmt.Errorf("unable to store fee reduction in redis %v", err)
	}
	return m, nil
}

// fetchFeeReduction determines the fee reduction based on token ownership
// when querying onchain
func (a *App) fetchFeeReduction(traderAddr string, chainId int) (float64, error) {
	c, exists := a.BrokerConf[int64(chainId)]
	if !exists {
		return 0, fmt.Errorf("no config for chain %d", chainId)
	}
	trdr := common.HexToAddress(traderAddr)
	multiplier := float64(1)
	for j := range c.RebateTokens {
		tkn := c.RebateTokens[j]
		var bal float64
		var err error
		for trial := 0; trial < 5; trial++ {
			bal, err = queryOnChainBalance(a.GlblRpc[int64(chainId)], int(tkn.Decimals), tkn.Address, trdr)
			if err != nil {
				slog.Info("attempt to query balance failed", "error", err)
				time.Sleep(2 * time.Second)
				continue
			}
			break
		}
		m := float64(1)
		if bal == 0 {
			continue
		}
		for k := range tkn.Scheme {
			if bal >= tkn.Scheme[k].Amount {
				m = tkn.Scheme[k].Multiplier
			}
		}
		multiplier = min(multiplier, m)
	}
	return multiplier, nil
}

func queryOnChainBalance(glblRpc *globalrpc.GlobalRpc, decimals int, tkn, trdr common.Address) (float64, error) {
	rec, err := glblRpc.GetAndLockRpc(globalrpc.TypeHTTPS, 10)
	defer glblRpc.ReturnLock(rec)
	if err != nil {
		return 0, fmt.Errorf("error getting GlblRPC %v", err)
	}
	client, err := ethclient.Dial(rec.Url)
	if err != nil {
		return 0, fmt.Errorf("error creating rpc %v", err)
	}
	bal, err := contracts.BalanceOf(client, int(decimals), tkn, trdr)
	if err != nil {
		return 0, fmt.Errorf("error querying balance %v", err)
	}
	return bal, nil
}

// GetVip3Level checks whether for the given address we already have
// vip3 level stored. if not, it will get it from the REST API and store
func (a *App) GetVip3Level(traderAddr string) int {
	c := *a.RedisClient.Client
	traderAddr = strings.ToLower(traderAddr)
	redisKey := VIP3_REDIS + ":" + traderAddr
	var lvl int
	lvlRedis, err := c.Do(context.Background(), c.B().Get().Key(redisKey).Build()).ToString()
	if err != nil {
		//query from rest
		lvl, err = RestGetVip3Level(traderAddr)
		if err != nil {
			slog.Error("Error in getting Vip3Level for trader addr " + traderAddr + ":" + err.Error())
			return 0
		}
		// store
		err = c.Do(context.Background(), c.B().Set().Key(redisKey).Value(strconv.Itoa(lvl)).Build()).Error()
		if err != nil {
			slog.Error("Error in stroring Vip3Level for trader addr " + traderAddr + ":" + err.Error())

		}
		// expire
		c.Do(context.Background(), c.B().Expire().Key(redisKey).Seconds(VIP3_INFO_EXPIRY_SEC).Build())
		return lvl
	}
	lvl, err = strconv.Atoi(lvlRedis)
	if err != nil {
		slog.Error("Error in getting Vip3Level for trader addr " + traderAddr + ":" + err.Error())
	}
	return lvl
}

// RestGetVip3Level queries the Vip3 level from the Vip3 API
func RestGetVip3Level(traderAddr string) (int, error) {
	query := API_URL + "?addr=" + traderAddr
	response, err := http.Get(query)
	if err != nil {
		return 0, errors.New("Error in GetVIP3Level:" + err.Error())
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, errors.New("Error in GetVIP3Level:" + err.Error())
	}
	var v Vip3Response
	err = json.Unmarshal(body, &v)
	if err != nil {
		return 0, errors.New("Error in GetVIP3Level:" + err.Error())
	}
	return v.Data.Level, nil
}

func vip3ToFeeMap(feeReduc string, brokerFeeTbps uint16) map[int][]uint16 {
	// parse string of the form "1101:70,70,70,70;196:70,70,70,70"
	if feeReduc == "" {
		return nil
	}
	chains := strings.Split(feeReduc, ";")
	reducedFees := make(map[int][]uint16)
	for _, c := range chains {
		v := strings.Split(c, ":")
		if len(v) == 1 {
			slog.Error("Error parsing VIP3 fees: provide chainId,e.g., 1101:70,70,70,70")
			return nil
		}
		chainId, err := strconv.Atoi(v[0])
		if err != nil {
			slog.Error("Error parsing VIP3 chainId to integer:" + err.Error())
			return nil
		}
		var fees []uint16

		feesStr := strings.Split(strings.TrimSuffix(v[1], ";"), ",")
		for k := 1; k < len(feesStr); k++ {
			valuePerc, err := strconv.Atoi(feesStr[k])
			if err != nil {
				slog.Error("Error converting VIP3 fee to integer:" + err.Error())
				return nil
			}
			fees = append(fees, uint16((100-valuePerc)*int(brokerFeeTbps)/100))
		}
		reducedFees[chainId] = fees
	}
	return reducedFees
}
