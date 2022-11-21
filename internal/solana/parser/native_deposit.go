package parser

import (
	"context"
	"database/sql"
	"fmt"
	tokentypes "gitlab.com/rarify-protocol/rarimo-core/x/tokenmanager/types"
	"gitlab.com/rarify-protocol/saver-grpc-lib/transactor"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/near/borsh-go"
	"github.com/olegfomenko/solana-go"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/config"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/data"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/data/pg"
	"gitlab.com/rarify-protocol/solana-program-go/contract"
)

type nativeParser struct {
	log     *logan.Entry
	storage *pg.Storage
	tx      transactor.Transactor
}

func NewNativeParser(cfg config.Config) *nativeParser {
	return &nativeParser{
		log:     cfg.Log(),
		storage: cfg.Storage(),
		tx:      cfg.Transactor(),
	}
}

var _ Parser = &nativeParser{}

func (n *nativeParser) ParseTransaction(tx solana.Signature, accounts []solana.PublicKey, instruction solana.CompiledInstruction, instructionId int) error {
	n.log.Infof("Found new native deposit in tx: %s id: %d", tx.String(), instructionId)
	var args contract.DepositNativeArgs

	err := borsh.Deserialize(&args, instruction.Data)
	if err != nil {
		return errors.Wrap(err, "error deserializing instruction data")
	}

	entry := &data.NativeDeposit{
		Hash:          tx.String(),
		InstructionID: instructionId,
		TargetNetwork: args.NetworkTo,

		Amount: int64(args.Amount),

		Receiver: args.ReceiverAddress,

		Sender: hexutil.Encode(accounts[contract.DepositNativeOwnerIndex].Bytes()),
	}

	if args.BundleData != nil && len(*args.BundleData) > 0 && args.BundleSeed != nil {
		entry.BundleData = sql.NullString{String: hexutil.Encode(*args.BundleData), Valid: true}
		entry.BundleData = sql.NullString{String: hexutil.Encode((*args.BundleSeed)[:]), Valid: true}
	}

	err = n.storage.Clone().NativeDepositQ().Insert(entry)
	if err != nil {
		return errors.Wrap(err, "error inserting native deposit", logan.F{
			"tx_hash": tx.String(),
		})
	}

	return n.tx.SubmitTransferOp(
		context.TODO(),
		hexutil.Encode(accounts[contract.DepositNativeOwnerIndex].Bytes()),
		tx.String(),
		fmt.Sprintf("%d", instructionId),
		args.NetworkTo,
		tokentypes.Type_NATIVE,
	)
}
