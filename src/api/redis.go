package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/redis/rueidis"
)

const (
	EXPIRATION_SEC        = 86_400 // fee reduction entries expire after a while
	REDIS_TRADER_FEE_MULT = "trdr_fee_mult"
)

type TraderFee struct {
	TimestampSec  int64   `json:"timestamp"`
	FeeMultiplier float64 `json:"feeMultiplier"`
}

// RedisStoreTraderFeeMultiplier stores the fee multiplier for a trader with an expiry
func RedisStoreTraderFeeMultiplier(client *rueidis.Client, chainId int, traderAddr string, feeMult float64) error {
	c := *client
	traderAddr = strings.ToLower(traderAddr)
	redisKey := REDIS_TRADER_FEE_MULT + ":" + strconv.Itoa(chainId) + ":" + traderAddr
	f := TraderFee{
		TimestampSec:  time.Now().Unix(),
		FeeMultiplier: feeMult,
	}
	j, err := json.Marshal(f)
	if err != nil {
		return err
	}
	cmd := c.B().Set().Key(redisKey).Value(string(j)).ExSeconds(EXPIRATION_SEC).Build()
	return c.Do(context.Background(), cmd).Error()
}

func RedisGetTraderFeeMultiplier(client *rueidis.Client, chainId int, traderAddr string) (TraderFee, error) {
	c := *client
	traderAddr = strings.ToLower(traderAddr)
	redisKey := REDIS_TRADER_FEE_MULT + ":" + strconv.Itoa(chainId) + ":" + traderAddr
	cmd := c.B().Get().Key(redisKey).Build()
	d, err := c.Do(context.Background(), cmd).ToString()
	if err != nil {
		return TraderFee{}, err
	}
	var f TraderFee
	err = json.Unmarshal([]byte(d), &f)
	if err != nil {
		slog.Error("unable to unmarshal data", "error", err)
		return TraderFee{}, err
	}
	return f, nil
}
