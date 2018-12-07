package model

import (
	"context"

	"github.com/copernet/whccommon/model"
	"github.com/jinzhu/gorm"
)

type PropertyHistoryExt struct {
	PropertyID int64
	TxID       int64
	TxType     uint64
	IsCreateTx int
}

func InsertPropertyHistory(model *model.PropertyHistory, ctx context.Context) {
	DB(ctx).Save(model)
}

func GetPropertyHistoryListSinceTxId(ctx context.Context, txId int64) ([]PropertyHistoryExt, error) {
	rows, err := DB(ctx).Raw("SELECT ph.property_id,ph.tx_id,txes.tx_type,find_in_set(txes.tx_type,'50,51,54') as is_create_tx from wormhole.property_histories ph join txes on ph.tx_id=txes.tx_id where ph.tx_id>=? order by ph.id asc",
		txId).Rows()
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	var item PropertyHistoryExt
	ret := make([]PropertyHistoryExt, 0)
	for rows.Next() {
		DB(ctx).ScanRows(rows, &item)
		ret = append(ret, item)
	}
	return ret, nil
}

func GetHistoryTxhashBeforeTxId(ctx context.Context, propertyId int64, txId int64) (*string, error) {
	var txHash string
	err := DB(ctx).Raw("SELECT txes.tx_hash from wormhole.property_histories ph join txes on ph.tx_id=txes.tx_id where ph.property_id=? and txes.tx_id<? order by txes.tx_id desc limit 1;", propertyId, txId).Row().Scan(&txHash)
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &txHash, nil
}

func DelPropertyHistory(ctx context.Context, propertyId int64, txId int64) error {
	return DB(ctx).Where("property_id=? and tx_id=?", propertyId, txId).Delete(&model.PropertyHistory{}).Error
}
