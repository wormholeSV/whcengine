package engine

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/copernet/whccommon/log"
	common "github.com/copernet/whccommon/model"
	"github.com/copernet/whcengine/model"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

func TestPath(t *testing.T) {
	CoreTask(log.NewContext())
}

func TestFor(t *testing.T) {
	for i := 1; i <= 1; i++ {
		fmt.Println(i)
	}

}

func TestInsertBlock(t *testing.T) {
	blockHash, _ := client.GetBlockHash(int64(543520))
	blockData, _ := client.GetBlockVerbose(blockHash)
	CheckPending(blockData, nil)
}

func TestListWHCBlock(t *testing.T) {

	for i := 1240767; i < 1252767; i++ {
		fmt.Println(i)
		list, _ := client.WhcListBlockTransactions(uint32(i))
		if len(list) > 0 {
			fmt.Println("end")
			fmt.Println(i)
			fmt.Println("end")
			break
		}
	}
}

func TestGetTransaction(t *testing.T) {
	tx := "309681e080507684c1abd3f37216d3a2dcb67d6a373c2c4c3c7841d2d97a0c7a"
	res, err := client.WhcGetTransaction(tx)
	if err != nil {
		t.Error(err.Error())
	}
	fmt.Println(res)
}

func TestGetAmount(t *testing.T) {
	amount := getAmount("0.00001000")
	fmt.Println(amount)

}

func TestWhcGetSto(t *testing.T) {
	filter := "*"
	res, err := client.WhcGetSto("bf3d30fc9c9424bdc6e38fc55320bad6cda9488e74296fc8dfb06cb2d9ee0fd9", &filter)
	fmt.Println(err.Error())
	fmt.Println(res)
}

func TestGetOverDue(t *testing.T) {
	ExpireCrowdSales(time.Now().Unix(), log.NewContext())
}

func TestGetUnMatureTx(t *testing.T) {
	BurnMatured(1255906, log.NewContext())
}

func TestLock(t *testing.T) {

	bool := model.TryLock("test22322", log.NewContext())
	fmt.Println(bool)
}

func TestGetSyncInfo(t *testing.T) {
	start, end, _ := getSyncInfo(log.NewContext())
	fmt.Println(start)
	fmt.Println(end)
}

func TestUpdateAddressInTxes(t *testing.T) {
	frozenTx := &common.AddressesInTx{ID: 6, BalanceAvailableCreditDebit: &decimal.Zero, BalanceFrozenCreditDebit: &decimal.Zero}
	err := model.UpdateAddressInTx(frozenTx, nil)
	fmt.Println(err.Error())
}

func TestTransaction(t *testing.T) {

	for blockHeight := 0; blockHeight <= 10; blockHeight++ {
		tx := model.BeginTransaction()
		//defer func() {
		//	if r := recover(); r != nil {
		//		tx.Rollback()
		//	}
		//}()
		//
		ctx := context.WithValue(log.NewContext(), "context", tx)
		ct := ctx.Value("context").(*gorm.DB)
		fmt.Println(ct)
		val := ctx.Value(log.DefaultTraceLabel)
		fmt.Println(val)
		if blockHeight == 1 {
			blockHeight = blockHeight - 1
			continue
		}

		fmt.Println(blockHeight)

	}
}

func TestFetchBlock(t *testing.T) {
	err := fetchBlockData(1265758, log.NewContext())
	fmt.Println(err)
}

func TestDecimal(t *testing.T) {
	d, _ := decimal.NewFromString("100.00010000")
	fmt.Println(d.Float64())

	base := make(map[string]interface{})
	base["totaltokens"] = "2"
	totalTokens := base["totaltokens"]
	if totalTokens != nil && totalTokens != "" {
		fmt.Println("test")
	}
	fmt.Println("false")
}

func TestInsertPending(t *testing.T)  {
	insertPending(log.NewContext())
}