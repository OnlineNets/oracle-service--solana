package listener

import (
	"context"
	"time"

	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	"github.com/olegfomenko/solana-go/rpc/ws"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/distributed_lab/running"
	"gitlab.com/rarimo/savers/sol-saver-svc/internal/config"
	"gitlab.com/rarimo/savers/sol-saver-svc/internal/service"
	"gitlab.com/rarimo/savers/sol-saver-svc/internal/service/saver"
)

const runnerName = "bridge-listener"

type Service struct {
	log       *logan.Entry
	processor *saver.TxProcessor
	solana    *rpc.Client

	programId  solana.PublicKey
	wsEndpoint string
}

func NewService(cfg config.Config) *Service {
	return &Service{
		log:        cfg.Log(),
		processor:  saver.NewTxProcessor(cfg),
		solana:     cfg.SolanaRPC(),
		programId:  cfg.ListenConf().ProgramId,
		wsEndpoint: cfg.SolanaWSEndpoint(),
	}
}

func (s *Service) Listen(ctx context.Context) {
	running.UntilSuccess(ctx, s.log, runnerName, s.listen, 5*time.Second, 5*time.Second)
}

func (s *Service) listen(ctx context.Context) (bool, error) {
	wsCtx, wsCancel := context.WithCancel(ctx)
	defer wsCancel()

	client, err := ws.Connect(wsCtx, s.wsEndpoint)
	if err != nil {
		return false, errors.Wrap(err, "error opening solana websocket")
	}

	sub, err := client.LogsSubscribeMentions(
		s.programId,
		rpc.CommitmentFinalized,
	)

	if err != nil {
		panic(errors.Wrap(err, "error subscribing to the program logs"))
	}

	defer sub.Unsubscribe()
	defer client.Close()

	for {
		select {
		case <-ctx.Done():
			return true, nil
		default:
			got, err := sub.Recv()
			if err != nil {
				return false, errors.Wrap(err, "failed to receive transaction")
			}

			tx, err := service.GetTransaction(ctx, s.solana, got.Value.Signature)
			if err != nil {
				s.log.WithError(err).Error("failed to get transaction " + got.Value.Signature.String())
				continue
			}

			if err = s.processor.ProcessTransaction(ctx, got.Value.Signature, tx); err != nil {
				s.log.WithError(err).Error("failed to process transaction " + got.Value.Signature.String())
			}
		}
	}
}
