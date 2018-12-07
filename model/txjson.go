package model

import (
	"context"
	"encoding/json"

	"github.com/copernet/whc.go/btcjson"
	common "github.com/copernet/whccommon/model"
)

func InsertTxJson(t *btcjson.GenerateTransactionResult, rawdata *string, txId int, ctx context.Context) {
	data, _ := json.Marshal(t)
	var model = &common.TxJson{
		TxID:     txId,
		Protocol: common.Wormhole,
		TxData:   string(data),
		RawData:  *rawdata,
	}
	DB(ctx).Save(model)
}
