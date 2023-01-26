package voter

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gogo/protobuf/proto"
	"github.com/near/borsh-go"
	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	rarimotypes "gitlab.com/rarimo/rarimo-core/x/rarimocore/types"
	tokentypes "gitlab.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/rarimo/savers/saver-grpc-lib/voter/verifiers"
	"gitlab.com/rarimo/solana-program-go/contract"
	"gitlab.com/rarimo/solana-program-go/metaplex"
	"google.golang.org/grpc"
)

type nftOperator struct {
	chain  string
	solana *rpc.Client
	rarimo *grpc.ClientConn
}

func (f *nftOperator) ParseTransaction(ctx context.Context, accounts []solana.PublicKey, instruction solana.CompiledInstruction, transfer *rarimotypes.Transfer) error {
	verifiedTransfer, err := f.GetOperation(ctx, accounts, instruction)
	if err != nil {
		return err
	}

	verifiedTransfer.Origin = transfer.Origin
	verifiedTransfer.Tx = transfer.Tx
	verifiedTransfer.EventId = transfer.EventId

	if proto.Equal(verifiedTransfer, transfer) {
		return verifiers.ErrWrongOperationContent
	}

	return nil
}

func (f *nftOperator) GetOperation(ctx context.Context, accounts []solana.PublicKey, instruction solana.CompiledInstruction) (*rarimotypes.Transfer, error) {
	var args contract.DepositNFTArgs
	if err := borsh.Deserialize(&args, instruction.Data); err != nil {
		return nil, err
	}

	tokenId := hexutil.Encode(accounts[contract.DepositNFTMintIndex].Bytes())
	address, err := f.getTokenCollectionAddress(accounts[contract.DepositNFTMintIndex])
	if err != nil {
		return nil, err
	}

	if address == "" {
		address = tokenId
	}

	from := &tokentypes.OnChainItemIndex{
		Chain:   f.chain,
		Address: address,
		TokenID: tokenId,
	}

	dataResp, err := tokentypes.NewQueryClient(f.rarimo).CollectionData(ctx, &tokentypes.QueryGetCollectionDataRequest{Chain: f.chain, Address: address})
	if err != nil {
		return nil, err
	}

	collectionResp, err := tokentypes.NewQueryClient(f.rarimo).Collection(ctx, &tokentypes.QueryGetCollectionRequest{Index: dataResp.Data.Collection})
	if err != nil {
		return nil, err
	}

	to := &tokentypes.OnChainItemIndex{
		Chain:   args.NetworkTo,
		Address: "",
		TokenID: "", // TODO
	}

	for _, index := range collectionResp.Collection.Data {
		if index.Chain == args.NetworkTo {
			to.Address = index.Address
			break
		}
	}

	if to.Address == "" {
		return nil, verifiers.ErrWrongOperationContent
	}

	var bundleData, bundleSeed string
	if args.BundleData != nil && len(*args.BundleData) > 0 && args.BundleSeed != nil {
		bundleData = hexutil.Encode(*args.BundleData)
		bundleSeed = hexutil.Encode((*args.BundleSeed)[:])
	}

	meta, err := f.getItemMeta(from)
	if err != nil {
		return nil, err
	}

	return &rarimotypes.Transfer{
		Receiver:   args.ReceiverAddress,
		Amount:     "1",
		BundleData: bundleData,
		BundleSalt: bundleSeed,
		From:       from,
		To:         to,
		Meta:       meta,
	}, nil
}

func (f *nftOperator) getItemMeta(from *tokentypes.OnChainItemIndex) (*tokentypes.ItemMetadata, error) {
	metadata, err := f.getMetadata(solana.MustPublicKeyFromBase58(from.TokenID))
	if err != nil {
		return nil, err
	}

	imageUrl, imageHash, err := verifiers.GetImage(metadata.Data.URI)
	if err != nil {
		return nil, err
	}

	return &tokentypes.ItemMetadata{
		ImageUri:  imageUrl,
		ImageHash: imageHash,
		Seed:      "", // TODO
		Name:      metadata.Data.Name,
		Symbol:    metadata.Data.Symbol,
		Uri:       metadata.Data.URI,
	}, nil
}

func (f *nftOperator) getMetadata(mint solana.PublicKey) (*metaplex.Metadata, error) {
	metadataAddress, _, err := solana.FindTokenMetadataAddress(mint)
	if err != nil {
		return nil, err
	}

	metadataInfo, err := f.solana.GetAccountInfo(context.TODO(), metadataAddress)
	if err != nil {
		return nil, err
	}

	metadata := new(metaplex.Metadata)
	return metadata, borsh.Deserialize(metadata, metadataInfo.Value.Data.GetBinary())
}

func (f *nftOperator) getTokenCollectionAddress(mint solana.PublicKey) (string, error) {
	metadata, err := f.getMetadata(mint)
	if err != nil {
		return "", err
	}

	// TODO maybe check for zero collection address instead of verified value
	if metadata.Collection == nil || metadata.Collection.Verified == false {
		return "", nil
	}

	return hexutil.Encode(metadata.Collection.Address.Bytes()), nil
}
