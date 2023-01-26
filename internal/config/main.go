package config

import (
	"github.com/olegfomenko/solana-go/rpc"
	"github.com/tendermint/tendermint/rpc/client/http"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/rarimo/savers/saver-grpc-lib/broadcaster"
	"google.golang.org/grpc"
)

type Config interface {
	comfig.Logger
	comfig.Listenerer
	pgdb.Databaser
	broadcaster.Broadcasterer

	Cosmos() *grpc.ClientConn
	Tendermint() *http.HTTP
	ListenConf() ListenConf
	SolanaRPC() *rpc.Client
	SolanaWSEndpoint() string
}

type config struct {
	comfig.Logger
	comfig.Listenerer
	pgdb.Databaser
	broadcaster.Broadcasterer

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
		Databaser:     pgdb.NewDatabaser(getter),
		Broadcasterer: broadcaster.New(getter),
	}
}
