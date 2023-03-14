package config

import (
	"github.com/olegfomenko/solana-go/rpc"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
)

func (c *config) SolanaRPC() *rpc.Client {
	return c.solRPC.Do(func() interface{} {
		var config struct {
			Url string `fig:"url"`
		}

		if err := figure.Out(&config).From(kv.MustGetStringMap(c.getter, "rpc")).Please(); err != nil {
			panic(err)
		}
		return rpc.New(config.Url)
	}).(*rpc.Client)
}

func (c *config) SolanaWSEndpoint() string {
	return c.solWS.Do(func() interface{} {
		var config struct {
			Url string `fig:"url"`
		}

		if err := figure.Out(&config).From(kv.MustGetStringMap(c.getter, "ws")).Please(); err != nil {
			panic(err)
		}
		return config.Url
	}).(string)
}
