package model

import (
	"context"

	common "github.com/copernet/whccommon/model"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type AddressInTxEx struct {
	*common.AddressesInTx
	TxHash string `json:"tx_hash" gorm:"type:varchar(64);not null"`
}

func InsertAddressInTx(model *common.AddressesInTx, ctx context.Context) {
	//Default Zero
	if model.BalanceFrozenCreditDebit == nil {
		model.BalanceFrozenCreditDebit = &decimal.Zero
	}

	DB(ctx).Save(model)
}

func UpdateAddressInTx(model *common.AddressesInTx, ctx context.Context) error {
	//Default Zero
	if model.BalanceFrozenCreditDebit == nil {
		model.BalanceFrozenCreditDebit = &decimal.Zero
	}

	return DB(ctx).Model(&common.AddressesInTx{}).Where("id=?", model.ID).Select("BalanceAvailableCreditDebit", "BalanceFrozenCreditDebit").Update(model).Error
}

func GetUnMatureTx(height int64) ([]AddressInTxEx, error) {
	rows, err := db.Raw("select intx.*,tx.tx_hash from addresses_in_txes intx join txes tx "+
		"on intx.tx_id = tx.tx_id where tx.tx_type = ? and tx.tx_state = ? and intx.address_role = ? and tx.tx_block_height=?",
		68, common.Valid, common.Buyer, height).Rows()
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	ret := make([]AddressInTxEx, 0)
	for rows.Next() {
		var item AddressInTxEx
		err = db.ScanRows(rows, &item)
		if err != nil {
			return nil, err
		}

		ret = append(ret, item)
	}

	return ret, nil
}

func GetAddressesInTxByTxId(ctx context.Context, txId int64) ([]common.AddressesInTx, error) {
	var rows []common.AddressesInTx
	err := DB(ctx).Where("tx_id = ?", txId).Find(&rows).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return rows, nil
}

func GetValidAddressesInTxByTxId(ctx context.Context, txId int64) ([]common.AddressesInTx, error) {
	var txFromDB common.Tx
	error := DB(ctx).Where("tx_id=? and tx_state=?", txId, common.Valid).Find(&txFromDB).Error
	if error == nil {
		return GetAddressesInTxByTxId(ctx, txId)
	}
	if gorm.IsRecordNotFoundError(error) {
		return nil, nil
	}
	return nil, error
}
