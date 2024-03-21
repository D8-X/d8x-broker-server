package utils

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/redis/rueidis"
)

type ChainConfig struct {
	ChainId           int64            `json:"chainId"`
	Name              string           `json:"name"`
	MultiPayCtrctAddr common.Address   `json:"multiPayCtrctAddr"`
	AllowedExecutors  []common.Address `json:"allowedExecutors"`
}

type RpcConfig struct {
	ChainId int64    `json:"chainId"`
	Rpc     []string `json:"HTTP"`
}

type APIBrokerOrderSignatureReq struct {
	Order     APIOrderSig `json:"order"`
	ChainId   int64       `json:"chainId"`
	Signature string      `json:"signature"`
}

func (req *APIBrokerOrderSignatureReq) CheckData() error {
	if req.ChainId == 0 {
		return fmt.Errorf("chainId not provided")
	}
	if req.Order.Deadline == 0 {
		return fmt.Errorf("request requires order with iDeadline")
	}
	zeroAddr := common.Address{}.Hex()
	if req.Order.TraderAddr == zeroAddr ||
		req.Order.TraderAddr == "" {
		return fmt.Errorf("order requires order with non-zero traderAddr")
	}
	if req.Order.PerpetualId == 0 {
		return fmt.Errorf("request requires order with iPerpetualId")
	}

	return nil
}

//	Required data for the broker-signature: \
//		iPerpetualId: number, brokerFeeTbps: number, traderAddr: string, iDeadline: number,
//
//		chainId: number, proxyAddress: string
type APIOrderSig struct {
	PerpetualId   int32  `json:"iPerpetualId"`  // broker sig
	BrokerFeeTbps uint16 `json:"brokerFeeTbps"` // broker sig
	BrokerAddr    string `json:"brokerAddr"`    // broker sig
	TraderAddr    string `json:"traderAddr"`    // broker sig
	Deadline      uint32 `json:"iDeadline"`     // broker sig
	// relevant for order digest
	Flags              uint32 `json:"flags"`
	FAmount            string `json:"fAmount"`
	FLimitPrice        string `json:"fLimitPrice"`
	FTriggerPrice      string `json:"fTriggerPrice"`
	LeverageTDR        uint16 `json:"leverageTDR"`
	BrokerSignature    []byte `json:"brokerSignature"`
	ExecutionTimestamp uint32 `json:"executionTimestamp"`
}

// Message from executor websocket on signing an order
type WSOrderResp struct {
	OrderId            string `json:"orderId"`
	TraderAddr         string `json:"traderAddr"`
	Deadline           uint32 `json:"iDeadline"`
	Flags              uint32 `json:"flags"`
	FAmount            string `json:"fAmount"`
	FLimitPrice        string `json:"fLimitPrice"`
	FTriggerPrice      string `json:"fTriggerPrice"`
	ExecutionTimestamp uint32 `json:"executionTimestamp"`
}

type APIBrokerSignatureRes struct {
	Order           APIOrderSig `json:"orderFields"`
	ChainId         int64       `json:"chainId"`
	BrokerSignature string      `json:"brokerSignature"`
	OrderDigest     string      `json:"orderDigest"`
	OrderId         string      `json:"orderId"`
}

type APIBrokerFeeRes struct {
	BrokerFeeTbps uint16
}

type RueidisClient struct {
	Client *rueidis.Client
	Ctx    context.Context
}

const CHANNEL_NEW_ORDER = "new-order"
const EXPIRY_HDATA_SEC = 120

// PubOrder stores the order in redis with the order id as key
func (r *RueidisClient) PubOrder(order APIOrderSig, orderId string, chainId int64) error {
	perpetualIdStr := strconv.Itoa(int(order.PerpetualId))
	chainIdStr := strconv.Itoa(int(chainId))
	err := (*r.Client).Do(r.Ctx, (*r.Client).B().Hset().Key(orderId).FieldValue().
		FieldValue("ChainId", chainIdStr).
		FieldValue("PerpetualId", perpetualIdStr).
		FieldValue("TraderAddr", order.TraderAddr).
		FieldValue("Deadline", strconv.Itoa(int(order.Deadline))).
		FieldValue("Flags", strconv.Itoa(int(order.Flags))).
		FieldValue("FAmount", order.FAmount).
		FieldValue("FLimitPrice", order.FLimitPrice).
		FieldValue("FTriggerPrice", order.FTriggerPrice).
		FieldValue("ExecutionTimestamp", strconv.Itoa(int(order.ExecutionTimestamp))).Build()).Error()
	if err != nil {
		return err
	}
	// set expiry of key
	(*r.Client).Do(r.Ctx, (*r.Client).B().Expire().Key(orderId).Seconds(EXPIRY_HDATA_SEC).Build())
	return nil
}

// OrderSubmission pushes the order id to the stack,
// and publishes a message
func (r *RueidisClient) OrderSubmission(orderIds []string) error {
	for _, orderId := range orderIds {
		// get order from redis
		hm, err := (*r.Client).Do(r.Ctx, (*r.Client).B().Hgetall().Key(orderId).Build()).AsStrMap()
		if err != nil {
			return errors.New("Could not get id " + orderId + ": " + err.Error())
		}
		if len(hm) == 0 {
			return errors.New("Could not find id " + orderId + " - expired or never submitted")
		}
		// add to stack
		stackName := hm["PerpetualId"] + ":" + hm["ChainId"]
		(*r.Client).Do(r.Ctx, (*r.Client).B().Lpush().Key(stackName).Element(orderId).Build())

		// publish message
		msg := stackName
		err = (*r.Client).Do(r.Ctx, (*r.Client).B().Publish().Channel(CHANNEL_NEW_ORDER).Message(msg).Build()).Error()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RueidisClient) Subscribe(channel string, fn func(msg rueidis.PubSubMessage)) error {
	client := (*r.Client)
	err := client.Receive(r.Ctx, client.B().Subscribe().Channel(channel).Build(), fn)
	return err
}
