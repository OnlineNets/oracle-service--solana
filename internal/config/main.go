package config

import (
	"github.com/olegfomenko/solana-go/rpc"
	"github.com/rarimo/saver-grpc-lib/broadcaster"
	"github.com/rarimo/saver-grpc-lib/metrics"
	"github.com/rarimo/saver-grpc-lib/voter"
	"github.com/tendermint/tendermint/rpc/client/http"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"google.golang.org/grpc"
)

type Config interface {
	comfig.Logger
	comfig.Listenerer
	broadcaster.Broadcasterer
	voter.Subscriberer
	metrics.Profilerer

	Cosmos() *grpc.ClientConn
	Tendermint() *http.HTTP
	ListenConf() ListenConf
	SolanaRPC() *rpc.Client
	SolanaWSEndpoint() string
}

type config struct {
	comfig.Logger
	comfig.Listenerer
	broadcaster.Broadcasterer
	voter.Subscriberer
	metrics.Profilerer

	cosmos     comfig.Once
	tendermint comfig.Once
	lconf      comfig.Once
	solRPC     comfig.Once
	solWS      comfig.Once

	getter kv.Getter
}

func New(getter kv.Getter) Config {
	return &config{
		getter:        getter,
		Logger:        comfig.NewLogger(getter, comfig.LoggerOpts{}),
		Listenerer:    comfig.NewListenerer(getter),
		Broadcasterer: broadcaster.New(getter),
		Subscriberer:  voter.NewSubscriberer(getter),
		Profilerer:    metrics.New(getter),
	}
}
