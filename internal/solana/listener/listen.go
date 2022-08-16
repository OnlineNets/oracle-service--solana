package listener

import (
	"context"

	"github.com/olegfomenko/solana-go/rpc"
	"github.com/olegfomenko/solana-go/rpc/ws"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/solana/contract"
)

func (l *listener) Listen(ctx context.Context) {
	wsCtx, wsCancel := context.WithCancel(ctx)
	defer wsCancel()

	client, err := ws.Connect(wsCtx, l.wsEndpoint)
	if err != nil {
		panic(errors.Wrap(err, "error opening solana websocket"))
	}

	sub, err := client.LogsSubscribeMentions(
		l.config.ProgramId,
		rpc.CommitmentFinalized,
	)

	if err != nil {
		panic(errors.Wrap(err, "error subscribing to the program logs"))
	}

	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			got, err := sub.Recv()
			if err != nil {
				panic(err)
			}

			tx, err := l.parser.GetTransaction(ctx, got.Value.Signature)
			if err != nil {
				l.log.WithError(err).Error("failed to get transaction " + got.Value.Signature.String())
				continue
			}

			err = l.parser.ParseTransaction(got.Value.Signature, tx, contract.InstructionDepositMetaplex, l.parseDepositMetaplex)
			if err != nil {
				l.log.WithError(err).Error("failed to process transaction " + got.Value.Signature.String())
			}
		}
	}
}
