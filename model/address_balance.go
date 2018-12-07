package model

import (
	"context"
	"encoding/json"

	"github.com/copernet/whccommon/log"
	common "github.com/copernet/whccommon/model"
	"github.com/jinzhu/gorm"
)

func GetAddressBalance(address string, protocol common.Protocol, pid int64, ctx context.Context) *common.AddressBalance {
	balance := &common.AddressBalance{}
	err := DB(ctx).Where("address=? and protocol=? and property_id=?", address, protocol, pid).First(&balance).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil
	}
	return balance
}

func InsertAddressBalance(model *common.AddressBalance, ctx context.Context) {
	writeLog(model, ctx)
	DB(ctx).Save(model)
}
func writeLog(model *common.AddressBalance, ctx context.Context) {
	data, _ := json.Marshal(model)
	log.WithCtx(ctx).Infof("balance_change:%s", string(data))
}

func UpdateAddressBalance(model *common.AddressBalance, ctx context.Context) {
	writeLog(model, ctx)
	DB(ctx).Where("address=? and protocol=? and property_id=? ", model.Address, model.Protocol, model.PropertyID).
		Model(&common.AddressBalance{}).
		Select("BalanceAvailable", "BalanceReserved", "BalanceAccepted", "BalanceFrozen","LastTxID").
		Update(model)
}

func UpdateAddressBalanceWithTx(model *common.AddressBalance, ctx context.Context) error {
	writeLog(model, ctx)
	return DB(ctx).Where("address=? and protocol=? and property_id=? ", model.Address, model.Protocol,
		model.PropertyID).Model(&common.AddressBalance{}).
		Select("BalanceAvailable", "BalanceReserved", "BalanceAccepted", "BalanceFrozen").
		Update(model).Error
}

func GetAddressBalanceWithTx(ctx context.Context, address string, protocol common.Protocol, pid int64) *common.AddressBalance {
	balance := &common.AddressBalance{}
	err := DB(ctx).Where("address=? and protocol=? and property_id=?", address, protocol, pid).First(&balance).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil
	}
	return balance
}
