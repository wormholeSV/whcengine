package model

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/copernet/whccommon/log"
	common "github.com/copernet/whccommon/model"
	"github.com/gomodule/redigo/redis"
)

func TestPushStack(t *testing.T) {

	vo := common.BalanceNotify{Address: "cc", PropertyID: 2, TxID: 100}
	bys, _ := json.Marshal(vo)
	b := PushStack(AddressBalanceTip, string(bys), log.NewContext())
	fmt.Println(b)
}

func TestPopStack(t *testing.T) {

	data, err := PopStack(AddressBalanceTip)

	var vo common.BalanceNotify
	json.Unmarshal([]byte(data),&vo)

	if err == redis.ErrNil {
		fmt.Println(vo)
	}
}
