package model

import (
	"strconv"
	"time"

	"github.com/copernet/whccommon/log"
	common "github.com/copernet/whccommon/model"
	"github.com/copernet/whcengine/config"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/net/context"
)

const (
	// update block tip
	UpdateBlockTip = "block:tip"
	// notify balance for wormhole
	BalanceUpdated = "balance:wormhole:updated"
	// notify balance for wormhole
	AddressBalanceTip = "address:balance:update"
	MempoolTxTip = "transaction:tip"
)

var RedisPool *redis.Pool

func init() {
	conf := config.GetConf()
	p, err := common.ConnectRedis(conf.Redis)
	if err != nil {
		panic(err)
	}

	RedisPool = p.Pool
}

func TryLock(key string, ctx context.Context) bool {
	ri := RedisPool.Get()
	defer ri.Close()

	res, err := redis.Int(ri.Do("SETNX", key, strconv.FormatInt(time.Now().Unix(), 10)))
	if err != nil {
		return false
	}

	if res == 1 {
		//set expire time,1 hour
		_, err := ri.Do("EXPIRE", key, 1200)
		if err != nil {
			return false
		}

		return true
	}

	return false
}

func ReleaseLock(key string, ctx context.Context) {
	ri := RedisPool.Get()
	defer ri.Close()
	ri.Do("DEl", key)
}

func PushStack(key string, address string, ctx context.Context) error {
	ri := RedisPool.Get()
	defer ri.Close()

	_, err := redis.Int(ri.Do("SADD", key, address))
	if err != nil {
		return err
	}

	//set expire time,1 hour
	_, err = ri.Do("EXPIRE", key, 120)
	if err != nil {
		return err
	}

	log.WithCtx(ctx).Infof("SADD key:%s,member:%s push to queue success", key, address)
	return nil
}

func PopStack(key string) (string, error) {
	ri := RedisPool.Get()
	defer ri.Close()

	res, err := redis.String(ri.Do("SPOP", key))
	if err != nil {
		return "", err
	}

	return res, nil
}
