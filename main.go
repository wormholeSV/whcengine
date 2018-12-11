package main

import (
	"time"

	"github.com/copernet/whccommon/log"
	"github.com/copernet/whcengine/config"
	"github.com/copernet/whcengine/logic"
)

func main() {
	log.InitLog(config.GetConf().Log)
	//Register Channel
	engine.Subscribe()

	ticker := time.NewTicker(time.Second * config.GetConf().Private.TickerSeconds)
	for t := range ticker.C {
		ctx := log.NewContext()
		log.WithCtx(ctx).Info("Ticker start:", t)
		engine.CoreTask(ctx)
	}

}
