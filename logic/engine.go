package engine

import (
	"context"
	"strconv"
	"time"

	"github.com/copernet/whc.go/btcjson"
	"github.com/copernet/whccommon/log"
	common "github.com/copernet/whccommon/model"
	"github.com/copernet/whcengine/config"
	"github.com/copernet/whcengine/model"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

var client = common.ConnRpc(config.GetConf().RPC)
var firstMPtxBlock = config.GetFirstBlockHeight()

func CoreTask(ctx context.Context) {
	//check lock
	if !model.TryLock(config.LockKey, ctx) {
		log.WithCtx(ctx).Info("Engine is running,please ignore.")
		return
	}

	//Regorg
	err := MaybeReorg(ctx)
	if err != nil {
		log.WithCtx(ctx).Errorf("MaybeReorg error:%s", err.Error())
		//releaseLock
		model.ReleaseLock(config.LockKey, ctx)
		return
	}

	//Schedule Insert Pending
	insertPending(ctx)

	start, end, err := getSyncInfo(ctx)
	if err != nil {
		log.WithCtx(ctx).Errorf("getSyncInfo error", err.Error())
		//releaseLock
		model.ReleaseLock(config.LockKey, ctx)
		return
	}

	log.WithCtx(ctx).Infof("[sync]startBlock:%d,endBlock:%d", start, end)

	for blockHeight := start; blockHeight <= end; blockHeight++ {
		start := time.Now()
		tx := model.BeginTransaction()
		ctxWithTrans := context.WithValue(ctx, model.TRANSACTION, tx)

		err := fetchBlockData(blockHeight, ctxWithTrans)
		if err != nil {
			log.WithCtx(ctx).Errorf("[Fetch]blockdata error:%s,Rollback BlockHeight:%d", err.Error(), blockHeight)

			tx = ctxWithTrans.Value(model.TRANSACTION).(*gorm.DB)
			tx.Rollback()
			blockHeight = blockHeight - 1
		}
		tx.Commit()
		end := time.Now()

		//logic for overdue crowd on chain,mark as invalid
		BurnMatured(int64(blockHeight), ctx)

		log.WithCtx(ctx).Infof("Fetch block:%d cost:%f", blockHeight, end.Sub(start).Seconds())
	}

	//Pub Msg for notify Balance Change
	notifyBalanceUpdate(ctx)
	//releaseLock
	model.ReleaseLock(config.LockKey, ctx)
}
func fetchBlockData(blockHeight int, ctx context.Context) error {
	blockHash, err := client.GetBlockHash(int64(blockHeight))
	if err != nil {
		return err
	}
	blockData, err := client.GetBlockVerbose(blockHash)
	if err != nil {
		return err
	}

	var blockDataMP []string
	height := blockData.Height
	if height >= firstMPtxBlock {
		blockDataMP, err = client.WhcListBlockTransactions(uint32(height))
		if err != nil {
			return err
		}
	}

	CheckPending(blockData, ctx)

	bits, _ := strconv.ParseUint(blockData.Bits, 16, 32)
	//insert block with blockHash for lock,if the lock wasn't complete fully,the next reorg will find it
	block := &common.Block{
		Version:     blockData.Version,
		BlockHeight: blockData.Height,
		BlockHash:   blockData.Hash,
		Nonce:       blockData.Nonce,
		Bits:        uint32(bits),
		PrevBlock:   blockData.PreviousHash,
		MerkleRoot:  blockData.MerkleRoot,
		BlockTime:   blockData.Time,
		Size:        blockData.Size,
		Txcount:     len(blockData.Tx),
		Whccount:    len(blockDataMP),
	}
	err = model.InsertBlock(block, ctx)
	if err != nil {
		return err
	}

	lastTx := model.GetLastTx(model.Desc, ctx)
	var lastTxId int
	if lastTx != nil {
		lastTxId = lastTx.TxID
	}

	x := 1
	if blockDataMP != nil {
		for _, tx := range blockDataMP {
			log.WithCtx(ctx).Infof("[engine]process txhash:%s", tx)

			transaction, err := client.WhcGetTransaction(tx)
			if err != nil {
				return err
			}

			lastTxId += 1
			err = insertTransaction(transaction, lastTxId, x, ctx)
			if err != nil {
				return err
			}
			x += 1
		}
	}

	err = ExpireCrowdSales(blockData.Time, ctx)
	if err != nil {
		return err
	}

	return nil
}

func getSyncInfo(ctx context.Context) (int, int, error) {
	currentBlock := model.GetLastBlock(ctx)

	start := config.GetFirstBlockHeight()
	if currentBlock != nil {
		start = currentBlock.BlockHeight + 1
	}

	info, err := client.GetInfo()
	if err != nil {
		return 0, 0, err
	}

	return int(start), int(info.Blocks), nil
}

func insertTxAddr(transaction *btcjson.GenerateTransactionResult, txid int, ctx context.Context) error {
	var amount *decimal.Decimal
	if transaction.TypeInt != 4 {
		amount = getAmount(transaction.Amount)
		if transaction.TypeInt == 1 && transaction.Valid {
			amount = getAmount(transaction.ActualInvested)
		}
	}

	handlers := BuildFactory(transaction.TypeInt)
	if handlers == nil {
		log.WithCtx(ctx).Warnf("Unknown Tx_type:%d", transaction.TypeInt)
		return nil
	}
	tx, err := handlers.Invoke(transaction, txid, amount, ctx)
	if err != nil {
		return err
	}

	if tx != nil {
		model.InsertAddressInTx(tx, ctx)
		if transaction.Valid {
			updateBalance(tx, ctx)
		}
	}

	return nil
}
func getAmount(amount string) *decimal.Decimal {
	value, _ := decimal.NewFromString(amount)
	return &value
}

/**
1、Found Pending TX, Age 5 hours,then remove
2、remove tx in blockData.tx
*/
func needRemove(tx common.Tx, blockData *btcjson.GetBlockVerboseResult) bool {
	removeOld := false
	if time.Now().Unix()-tx.CreatedAt.Unix() > 18000 {
		removeOld = true
	}

	for _, hash := range blockData.Tx {
		if hash == tx.TxHash {
			removeOld = true
			break
		}
	}

	return removeOld
}
