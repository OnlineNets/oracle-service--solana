package parser

import (
	"context"

	"github.com/near/borsh-go"
	pg_dao "github.com/olegfomenko/pg-dao"
	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/config"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/data"
	"gitlab.com/rarify-protocol/solana-program-go/contract"
	"gitlab.com/rarify-protocol/solana-program-go/metaplex"
)

type nftParser struct {
	log    *logan.Entry
	dao    pg_dao.DAO
	solana *rpc.Client
}

func NewNFTParser(cfg config.Config) *nftParser {
	return &nftParser{
		log:    cfg.Log(),
		dao:    pg_dao.NewDAO(cfg.DB(), data.NFTDepositsTableName),
		solana: cfg.SolanaRPC(),
	}
}

var _ Parser = &nftParser{}

func (f *nftParser) ParseTransaction(tx solana.Signature, accounts []solana.PublicKey, instruction solana.CompiledInstruction, instructionId uint32) error {
	f.log.Infof("Found new nft deposit in tx: %s id: %d", tx.String(), instructionId)
	var args contract.DepositNFTArgs

	err := borsh.Deserialize(&args, instruction.Data)
	if err != nil {
		return errors.Wrap(err, "error deserializing instruction data")
	}

	collection, err := f.getTokenCollectionAddress(accounts[contract.DepositNFTMintIndex])
	if err != nil {
		return errors.Wrap(err, "error getting token collection")
	}

	entry := data.NFTDeposit{
		Hash:          tx.String(),
		InstructionId: instructionId,
		Sender:        accounts[contract.DepositNFTOwnerIndex].String(),
		Receiver:      args.ReceiverAddress,
		TargetNetwork: args.NetworkTo,
		Mint:          accounts[contract.DepositNFTMintIndex].String(),
		Collection:    collection,
	}

	_, err = f.dao.Clone().Create(entry)
	return err
}

func (f *nftParser) getTokenCollectionAddress(mint solana.PublicKey) (string, error) {
	metadataAddress, _, err := solana.FindTokenMetadataAddress(mint)
	if err != nil {
		return "", err
	}

	metadataInfo, err := f.solana.GetAccountInfo(context.TODO(), metadataAddress)
	if err != nil {
		return "", err
	}

	var metadata metaplex.Metadata
	err = borsh.Deserialize(&metadata, metadataInfo.Value.Data.GetBinary())
	if err != nil {
		return "", err
	}

	if metadata.Collection == nil {
		return "", nil
	}

	return metadata.Collection.Address.String(), nil
}
