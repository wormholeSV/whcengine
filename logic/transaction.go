package engine

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"strconv"

	"github.com/bcext/gcash/chaincfg/chainhash"
	"github.com/copernet/whc.go/btcjson"
	"github.com/copernet/whccommon/log"
	common "github.com/copernet/whccommon/model"
	"github.com/copernet/whcengine/config"
	"github.com/copernet/whcengine/model"
	"github.com/shopspring/decimal"
)

func BurnMatured(height int64, ctx context.Context) error {

	blockHeight := height - 1000
	if config.GetConf().TestNet {
		blockHeight = height - 3
	}

	blockHeight = blockHeight + 1
	if blockHeight < firstMPtxBlock {
		return nil
	}

	//burn bch has matured
	frozenTxs, _ := model.GetUnMatureTx(blockHeight)
	for _, t := range frozenTxs {
		transaction, err := client.WhcGetTransaction(t.TxHash)
		if err != nil {
			return err
		}
		//UnMatured continue
		if !transaction.Mature {
			continue
		}

		//Matured,try assign FrozenCreditDebit to AvailableCreditDebit,empty FrozenCreditDebit
		frozenTx := &common.AddressesInTx{ID: t.ID, BalanceAvailableCreditDebit: t.BalanceFrozenCreditDebit, BalanceFrozenCreditDebit: &decimal.Zero}
		data, _ := json.Marshal(frozenTx)
		log.WithCtx(ctx).Infof("TxHash:%s has matured,update balance:%s", t.TxHash, string(data))
		model.UpdateAddressInTx(frozenTx, ctx)

		//updateBalance
		frozen := t.BalanceFrozenCreditDebit.Neg()
		tx := &common.AddressesInTx{Address: t.Address, PropertyID: 1, Protocol: common.Wormhole, TxID: -1,
			BalanceAvailableCreditDebit: t.BalanceFrozenCreditDebit, BalanceFrozenCreditDebit: &frozen}
		updateBalance(tx, ctx)
	}

	return nil
}

func ExpireCrowdSales(time int64, ctx context.Context) error {
	properties, err := model.GetOverdueCrowdSale(time, ctx)
	if err != nil {
		return err
	}

	for _, property := range properties {
		log.WithCtx(ctx).Infof("corwdSale pid:%d has expired", property.PropertyID)

		var result btcjson.WhcGetPropertyResult
		json.Unmarshal([]byte(property.PropertyData), &result)

		tx, err := client.WhcGetTransaction(result.CreateTxID)
		if err != nil {
			return err
		}

		err = upsertProperty(tx, ctx, false)
		if err != nil {
			return err
		}

		retrieveTokens(tx.PropertyID, 0, ctx)
	}

	return nil
}

func CheckPending(blockData *btcjson.GetBlockVerboseResult, ctx context.Context) {
	//check pending
	transactions := model.GetPendingTransaction(ctx)
	for _, tx := range transactions {
		if needRemove(tx, blockData) {
			model.RemoveTxRelation(tx.TxID, ctx)
		}
	}

}

func insertTransaction(transaction *btcjson.GenerateTransactionResult, lastTxId int, x int, ctx context.Context) error {
	hash, _ := chainhash.NewHashFromStr(transaction.TxID)
	txRet, err := client.GetRawTransaction(hash)
	buf := bytes.NewBuffer(make([]byte, 0, txRet.MsgTx().SerializeSize()))
	txRet.MsgTx().Serialize(buf)
	rawData := hex.EncodeToString(buf.Bytes())

	transaction.FeeRate = calcFeeRate(transaction.Fee, len(rawData)/2)

	//if dup record error,return
	lastTxId, err = model.InsertTx(transaction, lastTxId, x, ctx)
	if err != nil {
		return err
	}

	err = insertTxAddr(transaction, lastTxId, ctx)
	if err != nil {
		return err
	}
	model.InsertTxJson(transaction, &rawData, lastTxId, ctx)
	return nil
}

// fee: in satoshi
// size: in bytes
func calcFeeRate(fee string, size int) *decimal.Decimal {
	feeInt, _ := strconv.Atoi(fee)
	value := decimal.New(int64(feeInt), 0).
		Div(decimal.New(int64(size), 0)).
		Div(decimal.New(1e5, 0)).
		Truncate(8)

	return &value
}

func insertPending(ctx context.Context) {

	filter := ""
	pendTxs, _ := client.WhcListPendingTransactions(&filter)
	lastTx := model.GetLastTx(model.Asc, ctx)
	txId := -1
	if lastTx != nil && lastTx.TxID <= 0 {
		txId = lastTx.TxID - 1
	}

	for _, t := range pendTxs {
		log.WithCtx(ctx).Warnf("insertTransaction tx:%s", t.TxID)
		err := insertTransaction(&t, txId, 0, ctx)
		if err != nil {
			log.WithCtx(ctx).Warnf("Dup key tx:%s,error:%v", t.TxID, err)
		}

		txId = txId - 1
	}
}
