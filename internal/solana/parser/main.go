package parser

import (
	"context"

	bin "github.com/gagliardetto/binary"
	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/config"
	"gitlab.com/rarify-protocol/solana-program-go/contract"
)

const DataInstructionCodeIndex = 0

type Parser interface {
	ParseTransaction(tx solana.Signature, accounts []solana.PublicKey, instruction solana.CompiledInstruction, instructionId uint32) error
}

type Service struct {
	log     *logan.Entry
	solana  *rpc.Client
	program solana.PublicKey
	parsers map[contract.Instruction]Parser
}

func NewService(cfg config.Config) *Service {
	return &Service{
		log:     cfg.Log(),
		solana:  cfg.SolanaRPC(),
		program: cfg.ListenConf().ProgramId,
		parsers: map[contract.Instruction]Parser{
			contract.InstructionDepositNative: NewNativeParser(cfg),
			contract.InstructionDepositFT:     NewFTParser(cfg),
			contract.InstructionDepositNFT:    NewNFTParser(cfg),
		},
	}
}

// GetTransaction requests Solana transaction entry by signature
func (s *Service) GetTransaction(ctx context.Context, sig solana.Signature) (*solana.Transaction, error) {
	out, err := s.solana.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Encoding: solana.EncodingBase64,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error getting transaction from solana")
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(out.Transaction.GetBinary()))
	return tx, errors.Wrap(err, "error decoding transaction")
}

// ParseTransaction checks transaction program id and performs a parser call
// if the corresponding instruction parser present in parsers map
func (s *Service) ParseTransaction(sig solana.Signature, tx *solana.Transaction) error {
	accounts := tx.Message.AccountKeys
	s.log.Debug("Parsing transaction " + sig.String())

	for index, instruction := range tx.Message.Instructions {
		if accounts[instruction.ProgramIDIndex] == s.program {
			if parser, ok := s.parsers[contract.Instruction(instruction.Data[DataInstructionCodeIndex])]; ok {
				err := parser.ParseTransaction(sig, getInstructionAccounts(accounts, instruction.Accounts), instruction, uint32(index))
				if err != nil {
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
