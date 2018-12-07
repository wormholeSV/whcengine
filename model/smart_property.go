package model

import (
	"context"

	common "github.com/copernet/whccommon/model"
	"github.com/jinzhu/gorm"
)

func GetSmartProperty(pid int64, protocol common.Protocol, ctx context.Context) (*common.SmartProperty, error) {
	var sp = common.SmartProperty{}
	err := DB(ctx).Where("property_id=? and protocol=?", pid, protocol).First(&sp).Error
	return &sp, err
}

func UpsertSmartProperty(model *common.SmartProperty, ctx context.Context) {
	_, err := GetSmartProperty(model.PropertyID, common.Wormhole, ctx)
	if gorm.IsRecordNotFoundError(err) {
		DB(ctx).Save(model)
		return
	}

	DB(ctx).Model(&common.SmartProperty{}).Where("property_id=? and protocol=?", model.PropertyID, common.Wormhole).Select("Protocol", "PropertyData", "Issuer", "Ecosystem", "CreateTxID", "LastTxID", "PropertyName", "Precision", "PropertyCategory", "PropertySubcategory").Update(model)
}

func GetOverdueCrowdSale(time int64, ctx context.Context) ([]common.SmartProperty, error) {
	rows, err := DB(ctx).Raw("select property_id,property_data from smart_properties where property_data -> '$.active' = true and (property_data -> '$.deadline' < ? or property_data -> '$.endedtime' < ?)", time, time).Rows()
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	ret := make([]common.SmartProperty, 0)
	for rows.Next() {
		var item common.SmartProperty
		err = DB(ctx).ScanRows(rows, &item)
		if err != nil {
			return nil, err
		}

		ret = append(ret, item)
	}

	return ret, nil
}

func DelProperty(ctx context.Context, propertyId int64) (error) {
	return DB(ctx).Where("property_id=?", propertyId).Delete(&common.SmartProperty{}).Error
}
