package engine

import (
	"context"
	"encoding/json"

	"github.com/copernet/whc.go/btcjson"
	"github.com/copernet/whccommon/log"
	common "github.com/copernet/whccommon/model"
	"github.com/copernet/whccommon/mq"
	"github.com/copernet/whcengine/model"
)

var factory, _ = mq.BuildFactory(model.RedisPool)

const (
	Burn_Address = "bchtest:qqqqqqqqqqqqqqqqqqqqqqqqqqqqqdmwgvnjkt8whc"
)

func Subscribe() {
	ctx := log.NewContext()
	go factory.Subscribe(model.UpdateBlockTip, &NewBlockHandler{})
	log.WithCtx(ctx).Infof("Subscribe channel:%s", model.UpdateBlockTip)
	go factory.Subscribe(model.BalanceUpdated, &BalanceUpdateHandler{})
	log.WithCtx(ctx).Infof("Subscribe channel:%s", model.BalanceUpdated)
	go factory.Subscribe(model.MempoolTxTip, &TransactionTipHandler{})
	log.WithCtx(ctx).Infof("Subscribe channel:%s", model.MempoolTxTip)
}

type NewBlockHandler struct{}

func (t *NewBlockHandler) OnMessage(msg string, ctx context.Context) {
	log.WithCtx(ctx).Infof("Receive Message:%s", msg)
	CoreTask(ctx)
}

type BalanceUpdateHandler struct{}

func (t *BalanceUpdateHandler) OnMessage(msg string, ctx context.Context) {
	log.WithCtx(ctx).Infof("Receive Message:%s", msg)

	var notify common.BalanceNotify
	json.Unmarshal([]byte(msg), &notify)

	result, err := client.WhcGetAllBalancesForAddress(notify.Address)
	if err != nil {
		if e, ok := err.(*btcjson.RPCError); ok && (e.Code == -5 || e.Code == -8) {
			log.WithCtx(ctx).Warnf("WhcGetAllBalancesForAddress error:%s", err.Error())

		} else {
			log.WithCtx(ctx).Errorf("WhcGetAllBalancesForAddress error:%s", err.Error())
		}

		return
	}

	for _, vo := range result {
		balance := model.GetAddressBalance(notify.Address, common.Wormhole, int64(vo.PropertyID), ctx)

		if balance == nil {
			log.WithCtx(ctx).Errorf("[UnFind property]Address:%s,pid:%d", notify.Address, vo.PropertyID)
			continue
		}
		if !getAmount(vo.Balance).Equal(*balance.BalanceAvailable) {
			log.WithCtx(ctx).Errorf("[UnMatch Balance]Address:%s,pid:%d,core balance:%s,db balance:%s", notify.Address, vo.PropertyID, vo.Balance, (*balance.BalanceAvailable).String())
			continue
		}
	}
}

func notifyBalanceUpdate(ctx context.Context) {
	for {
		data, err := model.PopStack(model.AddressBalanceTip)
		if err != nil {
			break
		}

		var vo common.BalanceNotify
		json.Unmarshal([]byte(data), &vo)

		if vo.Address == Burn_Address {
			log.WithCtx(ctx).Infof("Ignore Burn_Address:%s", vo.Address)
			continue
		}

		log.WithCtx(ctx).Infof("pop msg:" + data)

		factory.Publish(model.BalanceUpdated, data, ctx)
	}

}

type TransactionTipHandler struct{}

func (t *TransactionTipHandler) OnMessage(msg string, ctx context.Context) {
	log.WithCtx(ctx).Infof("Receive Message:%s", msg)
	insertPending(ctx)

	log.WithCtx(ctx).Infof("Insert Pending ok:%s", msg)
}
