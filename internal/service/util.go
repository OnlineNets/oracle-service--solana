package service

import (
	"context"

	bin "github.com/gagliardetto/binary"
	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

func GetInstructionAccounts(accounts []solana.PublicKey, indexes []uint16) []solana.PublicKey {
	result := make([]solana.PublicKey, 0, len(indexes))
	for _, i := range indexes {
		result = append(result, accounts[i])
	}
	return result
}

// GetTransaction requests Solana transaction entry by signature.
// Returns <nil> if tx was not successful.
func GetTransaction(ctx context.Context, cli *rpc.Client, sig solana.Signature) (*solana.Transaction, error) {
	out, err := cli.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Encoding: solana.EncodingBase64,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error getting transaction from solana")
	}

	if out.Meta.Err != nil {
		return nil, nil
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(out.Transaction.GetBinary()))
	return tx, errors.Wrap(err, "error decoding transaction")
}
