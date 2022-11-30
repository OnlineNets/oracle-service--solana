package catchup

import (
	"context"
	"fmt"

	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/config"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/solana/parser"
)

type Service struct {
	log    *logan.Entry
	solana *rpc.Client
	parser *parser.Service

	programId solana.PublicKey
	fromTx    solana.Signature
	disabled  bool
}

func NewService(cfg config.Config) *Service {
	return &Service{
		log:    cfg.Log(),
		solana: cfg.SolanaRPC(),
		parser: parser.NewService(cfg),

		programId: cfg.ListenConf().ProgramId,
		fromTx:    cfg.ListenConf().FromTx,
		disabled:  cfg.ListenConf().DisableCatchup,
	}
}

// Catchup will list all transactions from last to specified in config and stored in l.fromTx
func (s *Service) Catchup(ctx context.Context) error {
	s.log.Info("Starting catchup")
	if s.disabled || s.fromTx.Equals(solana.Signature{}) {
		return nil
	}

	var start solana.Signature
	for {
		last, err := s.catchupFrom(ctx, start)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error catchupping history from %s", start))
		}

		if s.fromTx.Equals(last) {
			break
		}

		start = last
	}

	return nil
}

func (s *Service) catchupFrom(ctx context.Context, start solana.Signature) (solana.Signature, error) {
	s.log.Info(fmt.Sprintf("Catchupping history from %s", start))

	signatures, err := s.solana.GetSignaturesForAddressWithOpts(ctx, s.programId, &rpc.GetSignaturesForAddressOpts{
		Before:     start,
		Commitment: rpc.CommitmentFinalized,
	})

	if err != nil {
		return solana.Signature{}, errors.Wrap(err, "error getting txs")
	}

	for _, sig := range signatures {
		s.log.Debug("Checking tx: " + sig.Signature.String())
		tx, err := s.parser.GetTransaction(ctx, sig.Signature)
		if err != nil {
			s.log.WithError(err).Error("failed to get transaction " + sig.Signature.String())

			if s.fromTx.Equals(sig.Signature) {
				return sig.Signature, nil
			}

			continue
		}

		err = s.parser.ParseTransaction(sig.Signature, tx)
		if err != nil {
			s.log.WithError(err).Error("failed to process transaction " + sig.Signature.String())
		}

		if s.fromTx.Equals(sig.Signature) {
			return sig.Signature, nil
		}
	}

	if len(signatures) == 0 {
		return s.fromTx, nil
	}

	return signatures[len(signatures)-1].Signature, nil
}
