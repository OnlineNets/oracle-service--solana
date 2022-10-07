package config

import (
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/kit/pgdb"
)

type Config interface {
	comfig.Logger
	comfig.Listenerer
	pgdb.Databaser
	Solaner
	BridgeListener
	Storager
}

type config struct {
	comfig.Logger
	comfig.Listenerer
	pgdb.Databaser
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
	}
}
