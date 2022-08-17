package listener

import (
	"context"

	"github.com/near/borsh-go"
	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	"gitlab.com/distributed_lab/logan/v3/errors"
	lib "gitlab.com/rarify-protocol/saver-grpc-lib/grpc"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/data"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/solana/contract"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/solana/metaplex"
)

func (l *listener) parseDepositMetaplex(tx solana.Signature, accounts []solana.PublicKey, instruction solana.CompiledInstruction) error {
	l.log.Info("Found new deposit in tx: " + tx.String())
	var instructionData contract.DepositMetaplex

	err := borsh.Deserialize(&instructionData, instruction.Data)
	if err != nil {
		return errors.Wrap(err, "error deserializing instruction data")
	}

	entry := data.Transaction{
		Hash:          tx.String(),
		TokenMint:     accounts[contract.DepositMintIndex].String(),
		TargetNetwork: instructionData.NetworkTo,
		Receiver:      instructionData.ReceiverAddress,
		// TODO depends on contract call
		TokenType: int16(lib.Type_METAPLEX_NFT),
	}

	switch instructionData.TokenId {
	case nil:
		entry.TokenId = entry.TokenMint
	default:
		entry.TokenId = *instructionData.TokenId
	}

	collection, err := getTokenCollectionAddress(l.solana, accounts[contract.DepositMintIndex])
	if err != nil {
		return errors.Wrap(err, "error getting collection")
	}

	entry.Collection = collection.String()

	_, err = l.Transactions().Create(entry)
	return err
}

func getTokenCollectionAddress(rpc *rpc.Client, mint solana.PublicKey) (solana.PublicKey, error) {
	metadata, _, err := solana.FindTokenMetadataAddress(mint)
	if err != nil {
		return solana.PublicKey{}, err
	}

	metadataInfo, err := rpc.GetAccountInfo(context.TODO(), metadata)
	if err != nil {
		return solana.PublicKey{}, err
	}

	var data metaplex.Metadata
	err = borsh.Deserialize(&data, metadataInfo.Value.Data.GetBinary())
	if err != nil {
		return solana.PublicKey{}, err
	}

	if data.Collection == nil {
		return solana.PublicKey{}, nil
	}

	return data.Collection.Address, nil
}
