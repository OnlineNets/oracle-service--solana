package parser

import (
	"github.com/near/borsh-go"
	pg_dao "github.com/olegfomenko/pg-dao"
	"github.com/olegfomenko/solana-go"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/config"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/data"
	"gitlab.com/rarify-protocol/solana-program-go/contract"
)

type nativeParser struct {
	log *logan.Entry
	dao pg_dao.DAO
}

func NewNativeParser(cfg config.Config) *nativeParser {
	return &nativeParser{
		log: cfg.Log(),
		dao: pg_dao.NewDAO(cfg.DB(), data.NativeDepositsTableName),
	}
}

var _ Parser = &nativeParser{}

func (n *nativeParser) ParseTransaction(tx solana.Signature, accounts []solana.PublicKey, instruction solana.CompiledInstruction) error {
	n.log.Info("Found new native deposit in tx: " + tx.String())
	var args contract.DepositNativeArgs

	err := borsh.Deserialize(&args, instruction.Data)
	if err != nil {
		return errors.Wrap(err, "error deserializing instruction data")
	}

	entry := data.NativeDeposit{
		Id:            0,
		Hash:          tx.String(),
		Sender:        accounts[contract.DepositNativeOwnerIndex].String(),
		Receiver:      args.ReceiverAddress,
		TargetNetwork: args.NetworkTo,
		Amount:        args.Amount,
	}

	_, err = n.dao.Clone().Create(entry)
	return err
}
