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
)

const API_URL = "https://dappapi.vip3.io/api/v1/sbt/info"
const VIP3_REDIS = "VIP"
const VIP3_INFO_EXPIRY_SEC int64 = 7 * 86_400

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
	if a.BrokerFeeLvlsTbps == nil {
		return a.BrokerFeeTbps
	}
	return a.getReducedBrokerFeeTbps(traderAddr, chainId)
}

func (a *App) getReducedBrokerFeeTbps(traderAddr string, chainId int) uint16 {
	traderAddr = strings.ToLower(traderAddr)
	l := a.GetVip3Level(traderAddr)
	if l == 0 {
		return a.BrokerFeeTbps
	}
	if chainId == -1 {
		for key := range a.BrokerFeeLvlsTbps {
			chainId = key
			break
		}
		fmt.Printf("getReducedBrokerFeeTbps: chainId not specified, defaulting to %d", chainId)
	}
	if _, exits := a.BrokerFeeLvlsTbps[chainId]; !exits {
		fmt.Printf("getReducedBrokerFeeTbps: chainId %d queried but not specified", chainId)
		return a.BrokerFeeTbps
	}
	if l > len(a.BrokerFeeLvlsTbps[chainId]) {
		l = len(a.BrokerFeeLvlsTbps[chainId])
	}
	return a.BrokerFeeLvlsTbps[chainId][l-1]
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
