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

type ftParser struct {
	log *logan.Entry
	dao pg_dao.DAO
}

func NewFTParser(cfg config.Config) *ftParser {
	return &ftParser{
		log: cfg.Log(),
		dao: pg_dao.NewDAO(cfg.DB(), data.FTDepositsTableName),
	}
}

var _ Parser = &ftParser{}

func (f *ftParser) ParseTransaction(tx solana.Signature, accounts []solana.PublicKey, instruction solana.CompiledInstruction, instructionId uint32) error {
	f.log.Infof("Found new ft deposit in tx: %s id: %d", tx.String(), instructionId)
	var args contract.DepositFTArgs

	err := borsh.Deserialize(&args, instruction.Data)
	if err != nil {
		return errors.Wrap(err, "error deserializing instruction data")
	}

	entry := data.FTDeposit{
		Hash:          tx.String(),
		InstructionId: instructionId,
		Sender:        accounts[contract.DepositFTOwnerIndex].String(),
		Receiver:      args.ReceiverAddress,
		TargetNetwork: args.NetworkTo,
		Amount:        args.Amount,
		Mint:          accounts[contract.DepositFTMintIndex].String(),
	}
	_, err = f.dao.Clone().Create(entry)
	return err
}
