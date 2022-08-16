package listener

import (
	"context"

	pg_dao "github.com/olegfomenko/pg-dao"
	"github.com/olegfomenko/solana-go/rpc"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/config"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/data"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/solana/tx"
)

type Listener interface {
	Listen(ctx context.Context)
	Catchup(ctx context.Context) error
}

type listener struct {
	parser       *tx.Parser
	wsEndpoint   string
	solana       *rpc.Client
	config       config.ListenConf
	transactions pg_dao.DAO
	log          *logan.Entry
}

func NewListener(cfg config.Config) Listener {
	return &listener{
		parser:       tx.NewParser(cfg),
		wsEndpoint:   cfg.SolanaWSEndpoint(),
		solana:       cfg.SolanaRPC(),
		config:       cfg.ListenConf(),
		transactions: pg_dao.NewDAO(cfg.DB(), data.TransactionsTableName),
		log:          cfg.Log(),
	}
}

func (l *listener) Transactions() pg_dao.DAO {
	return l.transactions.Clone()
}
