package config

import (
	"time"

	"github.com/tendermint/tendermint/rpc/client/http"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func (c *config) Cosmos() *grpc.ClientConn {
	return c.cosmos.Do(func() interface{} {
		var config struct {
			Addr string `fig:"addr"`
		}

		if err := figure.Out(&config).From(kv.MustGetStringMap(c.getter, "cosmos")).Please(); err != nil {
			panic(err)
		}

		con, err := grpc.Dial(config.Addr, grpc.WithInsecure(), grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    10 * time.Second, // wait time before ping if no activity
			Timeout: 20 * time.Second, // ping timeout
		}))
		if err != nil {
			panic(err)
		}

		return con
	}).(*grpc.ClientConn)
}

func (c *config) Tendermint() *http.HTTP {
	return c.tendermint.Do(func() interface{} {
		var config struct {
			Addr string `fig:"addr"`
		}

		if err := figure.Out(&config).From(kv.MustGetStringMap(c.getter, "core")).Please(); err != nil {
			panic(err)
		}

		client, err := http.New(config.Addr, "/websocket")
		if err != nil {
			panic(err)
		}

		if err := client.Start(); err != nil {
			panic(err)
		}

		return client
	}).(*http.HTTP)
}
