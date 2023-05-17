package catchup

import (
	"context"
	"fmt"

	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarimo/savers/sol-saver-svc/internal/config"
	"gitlab.com/rarimo/savers/sol-saver-svc/internal/service"
	"gitlab.com/rarimo/savers/sol-saver-svc/internal/service/saver"
)

type Service struct {
	log       *logan.Entry
	solana    *rpc.Client
	processor *saver.TxProcessor

	programId solana.PublicKey
	fromTx    solana.Signature
}

func NewService(cfg config.Config) *Service {
	return &Service{
		log:       cfg.Log(),
		solana:    cfg.SolanaRPC(),
		processor: saver.NewTxProcessor(cfg),

		programId: cfg.ListenConf().ProgramId,
		fromTx:    cfg.ListenConf().FromTx,
	}
}

// Catchup will list all transactions from last to specified in config and stored in l.fromTx
func (s *Service) Catchup(ctx context.Context) error {
	s.log.Info("Starting catchup")
	if s.fromTx.Equals(solana.Signature{}) {
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
		tx, err := service.GetTransaction(ctx, s.solana, sig.Signature)
		if err != nil {
			s.log.WithError(err).Error("failed to get transaction " + sig.Signature.String())

			if s.fromTx.Equals(sig.Signature) {
				return sig.Signature, nil
			}

			continue
		}

		if tx == nil {
			continue
		}

		if err = s.processor.ProcessTransaction(ctx, sig.Signature, tx); err != nil {
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
