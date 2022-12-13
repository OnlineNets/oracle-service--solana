package config

import (
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/rarimo/savers/saver-grpc-lib/broadcaster"
)

type Config interface {
	comfig.Logger
	comfig.Listenerer
	pgdb.Databaser
	broadcaster.Broadcasterer
	Solaner
	BridgeListener
	Storager
}

type config struct {
	comfig.Logger
	comfig.Listenerer
	pgdb.Databaser
	broadcaster.Broadcasterer
	Solaner
	BridgeListener
	getter kv.Getter
	Storager
}

func New(getter kv.Getter) Config {
	return &config{
		getter:         getter,
		Logger:         comfig.NewLogger(getter, comfig.LoggerOpts{}),
		Listenerer:     comfig.NewListenerer(getter),
		Solaner:        NewSolaner(getter),
		BridgeListener: NewBridgeListener(getter),
		Storager:       NewStorager(getter),
		Databaser:      pgdb.NewDatabaser(getter),
		Broadcasterer:  broadcaster.New(getter),
	}
}
