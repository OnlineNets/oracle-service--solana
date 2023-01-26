package parser

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
	GetOperation(ctx context.Context, accounts []solana.PublicKey, instruction solana.CompiledInstruction) (*rarimotypes.Transfer, error)
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
				transfer, err := operator.GetOperation(context.TODO(), getInstructionAccounts(accounts, instruction.Accounts), instruction)
				if err != nil {
					return err
				}

				msg := &rarimotypes.MsgCreateTransferOp{
					Creator:    s.broadcaster.Sender(),
					Tx:         tx.String(),
					EventId:    fmt.Sprint(index),
					Receiver:   transfer.Receiver,
					Amount:     "", // TODO
					BundleData: transfer.BundleData,
					BundleSalt: transfer.BundleSalt,
					From:       transfer.From,
					To:         transfer.To,
					Meta:       transfer.Meta,
				}

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
