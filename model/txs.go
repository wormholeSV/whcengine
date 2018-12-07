package model

import (
	"context"
	"encoding/json"

	"github.com/copernet/whc.go/btcjson"
	"github.com/copernet/whccommon/log"
	"github.com/copernet/whccommon/model"
	common "github.com/copernet/whccommon/model"
	"github.com/jinzhu/gorm"
)

type OrderBy string

const (
	Asc  OrderBy = "asc"
	Desc OrderBy = "desc"
)

func GetPendingTransaction(ctx context.Context) []common.Tx {
	var txs []common.Tx
	DB(ctx).Where("tx_state=? and tx_id < 0", model.Pending).Find(&txs)
	return txs
}

/**
Remove tx relations by @Transaction
*/
func RemoveTxRelation(txId int, ctx context.Context) {
	DB(ctx).Where("tx_id=?", txId).Delete(&common.Tx{})
	DB(ctx).Where("tx_id=?", txId).Delete(&common.AddressesInTx{})
	DB(ctx).Where("tx_id=?", txId).Delete(&common.TxJson{})
	DB(ctx).Where("tx_id=?", txId).Delete(&common.PropertyHistory{})
}

func GetLastTx(orderBy OrderBy, ctx context.Context) *common.Tx {
	var tx = common.Tx{}
	var err error
	if orderBy == Asc {
		err = DB(ctx).Order("tx_id").Last(&tx).Error
	} else {
		err = DB(ctx).Order("tx_id desc").Last(&tx).Error
	}

	if gorm.IsRecordNotFoundError(err) {
		return nil
	}
	return &tx
}

func InsertTx(t *btcjson.GenerateTransactionResult, txId int, seqInBlock int, ctx context.Context) (int, error) {
	if txId < 0 {
		err := DB(ctx).Exec("INSERT INTO `txes` (`tx_hash`, `tx_id`, `protocol`, `tx_type`, `ecosystem`, `tx_state`, `block_time`, `tx_block_height`,`tx_seq_in_block`) "+
			"select ? as tx_hash, case when min(tx_id) < -1 then min(tx_id)-1 else -2 end as tx_id, ? as protocol, ? as tx_type, "+
			"? as ecosystem, ? as tx_state, ? as block_time ,? as tx_block_height,? as tx_seq_in_block from txes",
			t.TxID, common.Wormhole, t.TypeInt, common.Production, getTxState(t.Valid, txId), t.BlockTime, t.BlockHeight, seqInBlock).Error
		if err != nil {
			return 0, err
		}

		var tx common.Tx
		err = DB(ctx).Table("txes").Select("tx_id").Where("tx_hash = ? and protocol=?", t.TxID, common.Wormhole).Order("tx_id").First(&tx).Error
		return tx.TxID, err
	}

	var tx = &common.Tx{
		TxID:          txId,
		TxHash:        t.TxID,
		Protocol:      common.Wormhole,
		TxType:        t.TypeInt,
		Ecosystem:     common.Production,
		TxState:       getTxState(t.Valid, txId),
		TxBlockHeight: t.BlockHeight,
		TxSeqInBlock:  seqInBlock,
		BlockTime:     t.BlockTime,
	}
	return txId, DB(ctx).Save(tx).Error
}
func getTxState(vaid bool, txId int) common.TxState {
	if txId < 0 {
		return common.Pending
	}

	if vaid {
		return common.Valid
	} else {
		return common.InValid
	}
}

func GetTxByTxHash(txhash string, ctx context.Context) *common.Tx {
	var tx = common.Tx{}
	err := DB(ctx).Where("tx_hash=?", txhash).First(&tx).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil
	}
	return &tx
}

func GetTxidsSinceBlockHeight(ctx context.Context, blockHeight int64) ([]int64, error) {
	var txs []common.Tx
	err := DB(ctx).Order("id asc").Where("tx_block_height >= ?", blockHeight).Find(&txs).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	var txIds []int64
	for _, item := range txs {
		txIds = append(txIds, int64(item.TxID))
	}
	return txIds, nil
}

func RemoveTxRelationWithTransaction(ctx context.Context, txId int) error {
	if err := DB(ctx).Where("tx_id=?", txId).Delete(&common.Tx{}).Error; err != nil {
		return err
	}
	if err := DB(ctx).Where("tx_id=?", txId).Delete(&common.AddressesInTx{}).Error; err != nil {
		return err
	}
	if err := DB(ctx).Where("tx_id=?", txId).Delete(&common.TxJson{}).Error; err != nil {
		return err
	}
	return nil
}

func GetPropertyLastTxListBeforeHeight(ctx context.Context, needReorgFromHeight int64) []string {
	ret := make([]string, 0)
	txIds, _ := GetTxidsSinceBlockHeight(ctx, needReorgFromHeight)
	if txIds != nil && len(txIds) > 0 {
		minTxId := txIds[0]
		propertyHistoryRows, _ := GetPropertyHistoryListSinceTxId(ctx, minTxId)
		if propertyHistoryRows == nil {
			return ret
		}

		for _, ph := range propertyHistoryRows {
			if ph.IsCreateTx == 0 { // this ph's tx is not a createXXX tx,
				txhash, _ := GetHistoryTxhashBeforeTxId(ctx, ph.PropertyID, minTxId)
				if txhash != nil {
					ret = append(ret, *txhash)
				}
			}
		}
		return ret
	}
	return ret
}

func Reorg(needReorgFromHeight int64, needToUpdateProperties []common.SmartProperty, ctxOrigin context.Context) error {

	//1. find all the related blocks and delete them
	//2. deal with the properties and propertyHistorys.
	// 2.1 For the dirty blocks, if the properties are just created(tx's type is in(50,51,54)), delete the properties;
	// 2.2 else find the latest property's tx before Dirty block, use it to replay and set the property again
	//3. according to all the related and only valid addressesInTx(its related tx_id is a valid tx), undo the balance
	//4. find all the related txes, txjsons and addressesInTx, then delete them
	// above all are in a big transaction
	tx := db.Begin()
	ctx := context.WithValue(ctxOrigin, TRANSACTION, tx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return tx.Error
	}
	log.WithCtx(ctx).Infof("[reorg] step 1/4 : deleting affected blocks since height %d", needReorgFromHeight)
	if err := DeleteBlocksFromHeight(ctx, needReorgFromHeight); err != nil {
		tx.Rollback()
		return err
	}
	txIds, err := GetTxidsSinceBlockHeight(ctx, needReorgFromHeight)

	if err != nil {
		tx.Rollback()
		return err
	}
	log.WithCtx(ctx).Info("[reorg] step 2/4 : processing affected properties and propertyHistorys")
	if txIds != nil && len(txIds) > 0 {
		propertyHistoryRows, err := GetPropertyHistoryListSinceTxId(ctx, txIds[0])
		if err != nil {
			tx.Rollback()
			return err
		}

		for _, ph := range propertyHistoryRows {
			// a token is created in this propertyHistory's tx, so in the reorg cleaning part, we should del the smart_property record
			if ph.IsCreateTx > 0 {
				err := DelProperty(ctx, ph.PropertyID)
				if err != nil {
					tx.Rollback()
					return err
				}
			}
			err := DelPropertyHistory(ctx, ph.PropertyID, ph.TxID)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, property := range needToUpdateProperties {
			UpsertSmartProperty(&property, ctx)
		}
	}
	log.WithCtx(ctx).Info("[reorg] step 3/4 : undo the balance according to valid addressInTxes")
	log.WithCtx(ctx).Info("[reorg] step 4/4 : remove the tx,addressIntx,txJson records for every affected txId")
	for _, txId := range txIds {

		validAddressesInTxList, err := GetValidAddressesInTxByTxId(ctx, txId)
		if err != nil {
			tx.Rollback()
			return err
		}

		for _, addressesInTx := range validAddressesInTxList {
			if err := undoUpdateBalance(&addressesInTx, ctx); err != nil {
				tx.Rollback()
				return err
			}
		}

		if err := RemoveTxRelationWithTransaction(ctx, int(txId)); err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}
func undoUpdateBalance(tx *common.AddressesInTx, ctx context.Context) error {
	balance := GetAddressBalanceWithTx(ctx, tx.Address, tx.Protocol, tx.PropertyID)

	if balance == nil {
		//return errors.New("when undo the dirty txes, can't find corresponding balance which shouldn't happen")
		log.WithCtx(ctx).Warnf("when undo the dirty addressesInTxes, can't find corresponding balance which "+
			"shouldn't happen, the parameters are: address|%s, protocol|%s, propertyId|%s",
			tx.Address, tx.Protocol, tx.PropertyID)
		return nil
	}

	//undo balance data
	//Rest balance data
	if tx.BalanceAvailableCreditDebit != nil {
		*balance.BalanceAvailable = balance.BalanceAvailable.Sub(*tx.BalanceAvailableCreditDebit)
	}

	if tx.BalanceReservedCreditDebit != nil {
		*balance.BalanceReserved = balance.BalanceReserved.Sub(*tx.BalanceReservedCreditDebit)
	}

	if tx.BalanceAcceptedCreditDebit != nil {
		*balance.BalanceAccepted = balance.BalanceAccepted.Sub(*tx.BalanceAcceptedCreditDebit)
	}

	if tx.BalanceFrozenCreditDebit != nil {
		*balance.BalanceFrozen = balance.BalanceFrozen.Sub(*tx.BalanceFrozenCreditDebit)
	}

	if err := UpdateAddressBalanceWithTx(balance, ctx); err != nil {
		return err
	}

	vo := common.BalanceNotify{Address: tx.Address, PropertyID: tx.PropertyID, TxID: tx.TxID}
	bys, _ := json.Marshal(vo)
	//update balance in redis
	if err := PushStack(AddressBalanceTip, string(bys), ctx); err != nil {
		return err
	}

	return nil
}
