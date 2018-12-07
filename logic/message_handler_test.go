package engine

import (
	"testing"

	"github.com/copernet/whccommon/log"
)

func TestOnMessage(t *testing.T) {

	handler := &BalanceUpdateHandler{}
	handler.OnMessage("xxx", log.NewContext())
}
