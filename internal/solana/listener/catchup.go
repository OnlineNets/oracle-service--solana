package listener

import (
	"context"
	"fmt"
	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarify-protocol/solana-proxy-svc/internal/solana/contract"
)

// Catchup will list all transactions from last to specified in config and stored in l.fromTx
func (l *listener) Catchup(ctx context.Context) error {
	l.log.Info("Starting catchup")
	if l.config.DisableCatchup {
		return errors.New("catchup was disabled in config")
	}

	var start solana.Signature
	for {
		last, err := l.catchup(ctx, start)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error catchupping history from %s", start))
		}

		if l.config.FromTx.Equals(last) {
			break
		}

		start = last
	}

	return nil
}

func (l *listener) catchup(ctx context.Context, start solana.Signature) (solana.Signature, error) {
	l.log.Info(fmt.Sprintf("Catchupping history from %s", start))

	signatures, err := l.solana.GetSignaturesForAddressWithOpts(ctx, l.config.ProgramId, &rpc.GetSignaturesForAddressOpts{
		Before:     start,
		Commitment: rpc.CommitmentFinalized,
	})

	if err != nil {
		return solana.Signature{}, errors.Wrap(err, "error getting txs")
	}

	for _, sig := range signatures {
		l.log.Debug("Checking tx: " + sig.Signature.String())
		tx, err := l.parser.GetTransaction(ctx, sig.Signature)
		if err != nil {
			return solana.Signature{}, errors.Wrap(err, "failed to get transaction "+sig.Signature.String())
		}

		err = l.parser.ParseTransaction(sig.Signature, tx, contract.InstructionWithdrawMetaplex, l.parseWithdrawMetaplex)
		if err != nil {
			return solana.Signature{}, errors.Wrap(err, "failed to process transaction "+sig.Signature.String())
		}

		if l.config.FromTx.Equals(sig.Signature) {
			return sig.Signature, nil
		}
	}

	return signatures[len(signatures)-1].Signature, nil
}
