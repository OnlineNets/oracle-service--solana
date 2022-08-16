package config

import (
	"github.com/olegfomenko/solana-go/rpc"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type Solaner interface {
	SolanaRPC() *rpc.Client
	SolanaWSEndpoint() string
}

type solaner struct {
	getter kv.Getter
	rpc    comfig.Once
	ws     comfig.Once
}

func NewSolaner(getter kv.Getter) Solaner {
	return &solaner{
		getter: getter,
	}
}

func (s *solaner) SolanaRPC() *rpc.Client {
	return s.rpc.Do(func() interface{} {
		var config struct {
			Url string `fig:"url"`
		}

		if err := figure.Out(&config).From(kv.MustGetStringMap(s.getter, "rpc")).Please(); err != nil {
			panic(err)
		}
		return rpc.New(config.Url)
	}).(*rpc.Client)
}

func (s *solaner) SolanaWSEndpoint() string {
	return s.ws.Do(func() interface{} {
		var config struct {
			Url string `fig:"url"`
		}

		if err := figure.Out(&config).From(kv.MustGetStringMap(s.getter, "ws")).Please(); err != nil {
			panic(err)
		}
		return config.Url
	}).(string)
}
