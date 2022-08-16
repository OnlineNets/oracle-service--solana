package tx

import (
	"context"
	bin "github.com/gagliardetto/binary"
	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarify-protocol/solana-proxy-svc/internal/config"
	"gitlab.com/rarify-protocol/solana-proxy-svc/internal/solana/contract"
)

const DataInstructionCodeIndex = 0

type Parser struct {
	solana  *rpc.Client
	program solana.PublicKey
	log     *logan.Entry
}

func NewParser(cfg config.Config) *Parser {
	return &Parser{
		solana:  cfg.SolanaRPC(),
		program: cfg.ListenConf().ProgramId,
		log:     cfg.Log(),
	}
}

type ParserF func(tx solana.Signature, accounts []solana.PublicKey, instruction solana.CompiledInstruction) error

func (p *Parser) GetTransaction(ctx context.Context, sig solana.Signature) (*solana.Transaction, error) {
	out, err := p.solana.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Encoding: solana.EncodingBase64,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error getting transaction from solana")
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(out.Transaction.GetBinary()))
	return tx, errors.Wrap(err, "error decoding transaction")
}

func (p *Parser) ParseTransaction(sig solana.Signature, tx *solana.Transaction, inst contract.Instruction, f ParserF) error {
	accounts := tx.Message.AccountKeys
	p.log.Debug("Parsing transaction " + sig.String())

	for _, instruction := range tx.Message.Instructions {
		if accounts[instruction.ProgramIDIndex] == p.program {
			if instruction.Data[DataInstructionCodeIndex] == byte(inst) {
				return f(sig, getInstructionAccounts(accounts, instruction.Accounts), instruction)
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
