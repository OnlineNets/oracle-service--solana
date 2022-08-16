package listener

import (
	"context"
	"github.com/olegfomenko/solana-go/rpc"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/rarify-protocol/solana-proxy-svc/internal/config"
	"gitlab.com/rarify-protocol/solana-proxy-svc/internal/solana/tx"
)

type Listener interface {
	Listen(ctx context.Context)
	Catchup(ctx context.Context) error
}

type listener struct {
	parser     *tx.Parser
	wsEndpoint string
	solana     *rpc.Client
	config     config.ListenConf
	log        *logan.Entry
}

func NewListener(cfg config.Config) Listener {
	return &listener{
		parser:     tx.NewParser(cfg),
		wsEndpoint: cfg.SolanaWSEndpoint(),
		solana:     cfg.SolanaRPC(),
		config:     cfg.ListenConf(),
		log:        cfg.Log(),
	}
}
