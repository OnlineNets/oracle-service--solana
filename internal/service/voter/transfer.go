package voter

import (
	"context"
	"strconv"

	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	rarimotypes "gitlab.com/rarimo/rarimo-core/x/rarimocore/types"
	"gitlab.com/rarimo/savers/saver-grpc-lib/voter/verifiers"
	"gitlab.com/rarimo/savers/sol-saver-svc/internal/config"
	"gitlab.com/rarimo/savers/sol-saver-svc/internal/service"
	"gitlab.com/rarimo/solana-program-go/contract"
)

const DataInstructionCodeIndex = 0

type IOperator interface {
	ParseTransaction(ctx context.Context, accounts []solana.PublicKey, instruction solana.CompiledInstruction, transfer *rarimotypes.Transfer) error
}

type TransferOperator struct {
	solana    *rpc.Client
	program   solana.PublicKey
	chain     string
	operators map[contract.Instruction]IOperator
}

func NewTransferOperator(cfg config.Config) *TransferOperator {
	return &TransferOperator{
		program: cfg.ListenConf().ProgramId,
		chain:   cfg.ListenConf().Chain,
		operators: map[contract.Instruction]IOperator{
			contract.InstructionDepositNative: NewNativeOperator(cfg.ListenConf().Chain, cfg.Cosmos()),
			contract.InstructionDepositFT:     NewFTOperator(cfg.ListenConf().Chain, cfg.Cosmos()),
			contract.InstructionDepositNFT:    NewNFTOperator(cfg.ListenConf().Chain, cfg.SolanaRPC(), cfg.Cosmos()),
		},
	}
}

// Implements verifiers.ITransferOperator
var _ verifiers.TransferOperator = &TransferOperator{}

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

	transaction, err := service.GetTransaction(context.TODO(), t.solana, sig)
	if err != nil {
		return err
	}

	instruction := transaction.Message.Instructions[msgId]
	if transaction.Message.AccountKeys[instruction.ProgramIDIndex] == t.program {
		if operator, ok := t.operators[contract.Instruction(instruction.Data[DataInstructionCodeIndex])]; ok {
			return operator.ParseTransaction(context.TODO(), service.GetInstructionAccounts(transaction.Message.AccountKeys, instruction.Accounts), instruction, transfer)
		}
	}

	return verifiers.ErrWrongOperationContent
}
