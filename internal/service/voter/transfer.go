package voter

import (
	"context"
	"strconv"

	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	rarimotypes "github.com/rarimo/rarimo-core/x/rarimocore/types"
	"github.com/rarimo/saver-grpc-lib/voter/verifiers"
	"github.com/rarimo/sol-saver-svc/internal/config"
	"github.com/rarimo/sol-saver-svc/internal/service"
	"github.com/rarimo/solana-program-go/contracts/bridge"
)

const DataInstructionCodeIndex = 0

type IOperator interface {
	ParseTransaction(ctx context.Context, accounts []solana.PublicKey, instruction solana.CompiledInstruction, transfer *rarimotypes.Transfer) error
}

type TransferOperator struct {
	solana    *rpc.Client
	program   solana.PublicKey
	chain     string
	operators map[bridge.Instruction]IOperator
}

func NewTransferOperator(cfg config.Config) *TransferOperator {
	return &TransferOperator{
		solana:  cfg.SolanaRPC(),
		program: cfg.ListenConf().ProgramId,
		chain:   cfg.ListenConf().Chain,
		operators: map[bridge.Instruction]IOperator{
			bridge.InstructionDepositNative: NewNativeOperator(cfg.ListenConf().Chain, cfg.Log(), cfg.Cosmos()),
			bridge.InstructionDepositFT:     NewFTOperator(cfg.ListenConf().Chain, cfg.Log(), cfg.Cosmos()),
			bridge.InstructionDepositNFT:    NewNFTOperator(cfg.ListenConf().Chain, cfg.SolanaRPC(), cfg.Cosmos()),
		},
	}
}

// Implements verifiers.ITransferOperator
var _ verifiers.TransferOperator = &TransferOperator{}

func (t *TransferOperator) VerifyTransfer(ctx context.Context, tx, eventId string, transfer *rarimotypes.Transfer) error {
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

	transaction, err := service.GetTransaction(ctx, t.solana, sig)
	if err != nil {
		return err
	}

	instruction := transaction.Message.Instructions[msgId]

	if transaction.Message.AccountKeys[instruction.ProgramIDIndex] != t.program {
		return verifiers.ErrWrongOperationContent
	}

	operator, ok := t.operators[bridge.Instruction(instruction.Data[DataInstructionCodeIndex])]
	if !ok {
		return verifiers.ErrWrongOperationContent
	}

	return operator.ParseTransaction(ctx,
		service.GetInstructionAccounts(transaction.Message.AccountKeys, instruction.Accounts),
		instruction,
		transfer)
}
