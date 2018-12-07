package model

import (
	"context"

	common "github.com/copernet/whccommon/model"
	"github.com/jinzhu/gorm"
)

func GetLastBlock(ctx context.Context) *common.Block {
	var block = common.Block{}
	err := DB(ctx).Last(&block).Error

	if gorm.IsRecordNotFoundError(err) {
		return nil
	}

	return &block
}

func GetBlockByHeight(blockHeight int64) *common.Block {
	var block = common.Block{}
	db.Where("block_height=?", blockHeight).First(&block)
	return &block
}

func InsertBlock(block *common.Block, ctx context.Context) error {
	return DB(ctx).Save(block).Error
}

func DeleteBlocksFromHeight(ctx context.Context, blockHeight int64) error {
	return DB(ctx).Where("block_height >= ?", blockHeight).Delete(&common.Block{}).Error
}
