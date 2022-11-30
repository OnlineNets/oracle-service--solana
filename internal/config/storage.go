package config

import (
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/data/pg"
)

type Storager interface {
	Storage() *pg.Storage
}

type storager struct {
	getter kv.Getter
	once   comfig.Once
}

func NewStorager(getter kv.Getter) Storager {
	return &storager{
		getter: getter,
	}
}

func (s *storager) Storage() *pg.Storage {
	return s.once.Do(func() interface{} {
		return pg.New(pgdb.NewDatabaser(s.getter).DB())
	}).(*pg.Storage)
}
