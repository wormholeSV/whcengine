package model

import (
	"context"

	common "github.com/copernet/whccommon/model"
	"github.com/copernet/whcengine/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var db *gorm.DB

func init() {
	var err error
	db, err = common.ConnectDatabase(config.GetConf().DB)
	if err != nil {
		panic("initial database error!")
	}
}

func BeginTransaction() *gorm.DB {
	return db.Begin()
}

func DB(ctx context.Context) *gorm.DB {
	if ctx == nil || ctx.Value(TRANSACTION) == nil {
		return db
	}

	return ctx.Value(TRANSACTION).(*gorm.DB)
}
