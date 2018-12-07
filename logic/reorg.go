package engine

import (
	"context"
	"errors"

	"github.com/copernet/whccommon/log"
	common "github.com/copernet/whccommon/model"
	"github.com/copernet/whcengine/model"
)

func getReorgCheckRange(ctx context.Context) (int64, int64, error) {
	var fromHeight, toHeight int64
	lastBlock := model.GetLastBlock(ctx)
	if lastBlock == nil {
		return firstMPtxBlock, firstMPtxBlock, nil
	}
	toHeight = lastBlock.BlockHeight
	fromHeight = toHeight - 10
	if firstMPtxBlock <= fromHeight {
		return fromHeight, toHeight, nil
	}
	if firstMPtxBlock < toHeight {
		return firstMPtxBlock, toHeight, nil
	}
	return firstMPtxBlock, firstMPtxBlock, nil
}

func MaybeReorg(ctx context.Context) error {
	fromHeight, toHeight, err := getReorgCheckRange(ctx)
	// no error return
	if err != nil {
		return errors.New("db error when check whether needs reorg " + err.Error())
	}
	log.WithCtx(ctx).Infof("[reorg]check height from [%d] to [%d], make sure the db nodes are synced well with core nodes", fromHeight, toHeight)
	if fromHeight == toHeight {
		return nil
	}
	var needReorgFromHeight int64 = 0
	for blockHeight := fromHeight; blockHeight <= toHeight; blockHeight++ {
		hashFromClient, err := client.GetBlockHash(blockHeight)
		if err != nil {
			return err
		}

		hashFromDB := model.GetBlockByHeight(blockHeight).BlockHash
		if hashFromClient.String() != hashFromDB {
			needReorgFromHeight = blockHeight
			log.WithCtx(ctx).Infof("[reorg]find collision, need to do reorg since the height [%d],localBlockHash [%s], remoteBlockHash [%s]", needReorgFromHeight, hashFromDB, hashFromClient.String())
			break
		}
	}
	if needReorgFromHeight != 0 {
		txHashes := model.GetPropertyLastTxListBeforeHeight(ctx, needReorgFromHeight)
		sp := make([]common.SmartProperty, 0)
		for _, txHash := range txHashes {
			t, _ := client.WhcGetTransaction(txHash)
			property, _, _ := FetchPropertyAndHistory(t, ctx, false)
			if property != nil {
				sp = append(sp, *property)
			}
		}
		return model.Reorg(needReorgFromHeight, sp, ctx)

	}
	return nil
}
