package saver

import (
	"context"
	"fmt"

	"github.com/olegfomenko/solana-go"
	"gitlab.com/distributed_lab/logan/v3"
	rarimotypes "gitlab.com/rarimo/rarimo-core/x/rarimocore/types"
	"gitlab.com/rarimo/savers/saver-grpc-lib/broadcaster"
	"gitlab.com/rarimo/solana-program-go/contract"
)

const (
	DataInstructionCodeIndex = 0
)

type IOperator interface {
	GetMessage(ctx context.Context, accounts []solana.PublicKey, instruction solana.CompiledInstruction) (*rarimotypes.MsgCreateTransferOp, error)
}

type Service struct {
	log         *logan.Entry
	program     solana.PublicKey
	operators   map[contract.Instruction]IOperator
	broadcaster broadcaster.Broadcaster
}

// ParseTransaction checks transaction program id and performs a parser call
// if the corresponding instruction parser present in parsers map
func (s *Service) ParseTransaction(sig solana.Signature, tx *solana.Transaction) error {
	accounts := tx.Message.AccountKeys
	s.log.Debug("Parsing transaction " + sig.String())

	for index, instruction := range tx.Message.Instructions {
		if accounts[instruction.ProgramIDIndex] == s.program {
			if operator, ok := s.operators[contract.Instruction(instruction.Data[DataInstructionCodeIndex])]; ok {
				msg, err := operator.GetMessage(context.TODO(), getInstructionAccounts(accounts, instruction.Accounts), instruction)
				if err != nil {
					return err
				}

				msg.Creator = s.broadcaster.Sender()
				msg.Tx = sig.String()
				msg.EventId = fmt.Sprint(index)

				if err := s.broadcaster.BroadcastTx(context.TODO(), msg); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func getInstructionAccounts(accounts []solana.PublicKey, indexes []uint16) []solana.PublicKey {
	result := make([]solana.PublicKey, 0, len(indexes))
	for _, i := range indexes {
		result = append(result, accounts[i])
	}
	return result
}
