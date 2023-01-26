package voter

import (
	"context"
	"strconv"

	bin "github.com/gagliardetto/binary"
	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	rarimotypes "gitlab.com/rarimo/rarimo-core/x/rarimocore/types"
	"gitlab.com/rarimo/savers/saver-grpc-lib/voter/verifiers"
	"gitlab.com/rarimo/solana-program-go/contract"
)

const DataInstructionCodeIndex = 0

type IOperator interface {
	ParseTransaction(ctx context.Context, accounts []solana.PublicKey, instruction solana.CompiledInstruction, transfer *rarimotypes.Transfer) error
}

type TransferOperator struct {
	log       *logan.Entry
	solana    *rpc.Client
	program   solana.PublicKey
	chain     string
	operators map[contract.Instruction]IOperator
}

// Implements verifiers.ITransferOperator
var _ verifiers.ITransferOperator = &TransferOperator{}

func (t *TransferOperator) VerifyTransfer(tx, eventId string, transfer *rarimotypes.Transfer) error {
	if transfer.From.Chain != t.chain {
		return verifiers.ErrUnsupportedNetwork
	}

	sig, err := solana.SignatureFromBase58(tx)
	if err != nil {
		return err
	}

	msgId, err := strconv.Atoi(eventId)
	if err != nil {
		return err
	}

	transaction, err := t.GetTransaction(context.TODO(), sig)
	if err != nil {
		return err
	}

	instruction := transaction.Message.Instructions[msgId]
	if transaction.Message.AccountKeys[instruction.ProgramIDIndex] == t.program {
		if operator, ok := t.operators[contract.Instruction(instruction.Data[DataInstructionCodeIndex])]; ok {
			return operator.ParseTransaction(context.TODO(), getInstructionAccounts(transaction.Message.AccountKeys, instruction.Accounts), instruction, transfer)
		}
	}

	return verifiers.ErrWrongOperationContent
}

// GetTransaction requests Solana transaction entry by signature
func (t *TransferOperator) GetTransaction(ctx context.Context, sig solana.Signature) (*solana.Transaction, error) {
	out, err := t.solana.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Encoding: solana.EncodingBase64,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error getting transaction from solana")
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(out.Transaction.GetBinary()))
	return tx, errors.Wrap(err, "error decoding transaction")
}
