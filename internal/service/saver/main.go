package saver

import (
	"context"
	"fmt"

	"github.com/olegfomenko/solana-go"
	oracletypes "github.com/rarimo/rarimo-core/x/oraclemanager/types"
	"github.com/rarimo/saver-grpc-lib/broadcaster"
	"github.com/rarimo/sol-saver-svc/internal/config"
	"github.com/rarimo/sol-saver-svc/internal/service"
	"github.com/rarimo/sol-saver-svc/internal/service/voter"
	"github.com/rarimo/solana-program-go/contracts/bridge"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

const (
	DataInstructionCodeIndex = 0
)

type IOperator interface {
	GetMessage(ctx context.Context, accounts []solana.PublicKey, instruction solana.CompiledInstruction) (*oracletypes.MsgCreateTransferOp, error)
}

type TxProcessor struct {
	log         *logan.Entry
	program     solana.PublicKey
	operators   map[bridge.Instruction]IOperator
	broadcaster broadcaster.Broadcaster
}

func NewTxProcessor(cfg config.Config) *TxProcessor {
	return &TxProcessor{
		log:         cfg.Log(),
		program:     cfg.ListenConf().ProgramId,
		broadcaster: cfg.Broadcaster(),
		operators: map[bridge.Instruction]IOperator{
			bridge.InstructionDepositNative: voter.NewNativeOperator(cfg.ListenConf().Chain, cfg.Log(), cfg.Cosmos()),
			bridge.InstructionDepositFT:     voter.NewFTOperator(cfg.ListenConf().Chain, cfg.Log(), cfg.Cosmos()),
			bridge.InstructionDepositNFT:    voter.NewNFTOperator(cfg.ListenConf().Chain, cfg.SolanaRPC(), cfg.Cosmos()),
		},
	}
}

func (s *TxProcessor) ProcessTransaction(ctx context.Context, sig solana.Signature, tx *solana.Transaction) error {
	accounts := tx.Message.AccountKeys
	s.log.Debug("Parsing transaction " + sig.String())

	for index, instruction := range tx.Message.Instructions {
		if accounts[instruction.ProgramIDIndex] == s.program {
			if operator, ok := s.operators[bridge.Instruction(instruction.Data[DataInstructionCodeIndex])]; ok {
				msg, err := operator.GetMessage(ctx, service.GetInstructionAccounts(accounts, instruction.Accounts), instruction)
				if err != nil {
					return errors.Wrap(err, "error getting message")
				}

				msg.Creator = s.broadcaster.Sender()
				msg.Tx = sig.String()
				msg.EventId = fmt.Sprint(index)

				if err := s.broadcaster.BroadcastTx(ctx, msg); err != nil {
					return errors.Wrap(err, "error broadcasting tx")
				}
			}
		}
	}

	return nil
}
