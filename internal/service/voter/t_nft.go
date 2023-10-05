package voter

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gogo/protobuf/proto"
	"github.com/near/borsh-go"
	"github.com/olegfomenko/solana-go"
	"github.com/olegfomenko/solana-go/rpc"
	oracletypes "github.com/rarimo/rarimo-core/x/oraclemanager/types"
	rarimotypes "github.com/rarimo/rarimo-core/x/rarimocore/types"
	tokentypes "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"github.com/rarimo/saver-grpc-lib/voter/verifiers"
	"github.com/rarimo/solana-program-go/contracts/bridge"
	"github.com/rarimo/solana-program-go/metaplex"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type nftOperator struct {
	chain  string
	solana *rpc.Client
	rarimo *grpc.ClientConn
}

func NewNFTOperator(chain string, solana *rpc.Client, rarimo *grpc.ClientConn) *nftOperator {
	return &nftOperator{
		chain:  chain,
		solana: solana,
		rarimo: rarimo,
	}
}

func (f *nftOperator) ParseTransaction(ctx context.Context, accounts []solana.PublicKey, instruction solana.CompiledInstruction, transfer *rarimotypes.Transfer) error {
	msg, err := f.GetMessage(ctx, accounts, instruction)
	if err != nil {
		return errors.Wrap(err, "error getting message")
	}

	msg.Tx = transfer.Tx
	msg.EventId = transfer.EventId

	transferResp, err := oracletypes.NewQueryClient(f.rarimo).Transfer(ctx, &oracletypes.QueryGetTransferRequest{Msg: *msg})
	if err != nil {
		return errors.Wrap(err, "error querying transfer from core")
	}

	if !proto.Equal(&transferResp.Transfer, transfer) {
		return verifiers.ErrWrongOperationContent
	}

	return nil
}

func (f *nftOperator) GetMessage(ctx context.Context, accounts []solana.PublicKey, instruction solana.CompiledInstruction) (*oracletypes.MsgCreateTransferOp, error) {
	var args bridge.DepositNFTArgs
	if err := borsh.Deserialize(&args, instruction.Data); err != nil {
		return nil, errors.Wrap(err, "error desser tx args")
	}

	tokenId := hexutil.Encode(accounts[bridge.DepositNFTMintIndex].Bytes())
	address, err := f.getTokenCollectionAddress(accounts[bridge.DepositNFTMintIndex])
	if err != nil {
		return nil, errors.Wrap(err, "error getting collection address")
	}

	if address == "" {
		address = tokenId
	}

	from := tokentypes.OnChainItemIndex{
		Chain:   f.chain,
		Address: address,
		TokenID: tokenId,
	}

	to, err := f.getTargetOnChainItem(ctx, &from, args.NetworkTo)
	if err != nil {
		return nil, err
	}

	meta, err := f.getItemMeta(ctx, &from)
	if err != nil {
		return nil, err
	}

	msg := &oracletypes.MsgCreateTransferOp{
		Receiver: args.ReceiverAddress,
		Sender:   accounts[bridge.DepositNFTOwnerIndex].String(),
		Amount:   "1",
		From:     from,
		To:       *to,
		Meta:     meta,
	}

	if args.BundleData != nil && len(*args.BundleData) > 0 && args.BundleSeed != nil {
		msg.BundleData = hexutil.Encode(*args.BundleData)
		msg.BundleSalt = hexutil.Encode((*args.BundleSeed)[:])
	}

	return msg, nil
}

// getTargetOnChainItem generates target OnChainItem based on current item information and its native mint information
// If target exists => use its data
// If target chain is Solana => target id will be equal to the current one
// If native chain is Solana => target id will be equal to the current one
// If native chain is EVM => target id will be equal to the EVM
// If native chain is Near => target id will be equal to the hex(near id str)
func (f *nftOperator) getTargetOnChainItem(ctx context.Context, from *tokentypes.OnChainItemIndex, toChain string) (*tokentypes.OnChainItemIndex, error) {
	// 1. checking corner cases (from == to)
	if from.Chain == toChain {
		return from, nil
	}

	// 2. trying to check the existence of target OnChainItem
	to, err := f.tryGetOnChainItem(ctx, from, toChain)
	if err != nil {
		return nil, errors.Wrap(err, "unexpected error during fetching on chain item")
	}

	// 3. if exists - return it
	if to != nil {
		return to, nil
	}

	// 4. getting target data (should exist)
	targetDataIndex, err := f.getTargetDataIndex(ctx, from, toChain)
	if err != nil {
		return nil, errors.Wrap(err, "unexpected error during fetching target collection data index")
	}

	// 5. getting native collection data (should exist)
	nativeCollectionData, err := f.getNativeData(ctx, from)
	if err != nil {
		return nil, errors.Wrap(err, "unexpected error during fetching native collection data")
	}

	// 6. If its equal to the current chain
	if nativeCollectionData.Index.Chain == from.Chain {
		return &tokentypes.OnChainItemIndex{
			Chain:   targetDataIndex.Chain,
			Address: targetDataIndex.Address,
			TokenID: from.TokenID,
		}, nil
	}

	// 7. getting native OnChainItem (should exist)
	native, err := f.tryGetOnChainItem(ctx, from, nativeCollectionData.Index.Chain)
	if err != nil {
		return nil, errors.Wrap(err, "unexpected error during fetching native onChainItem")
	}

	if native == nil {
		return nil, verifiers.ErrWrongOperationContent
	}

	// TODO manage several solana networks supported
	// 8. then target token id is equal to the native token id (in case of Near chain it already store in hex format)
	return &tokentypes.OnChainItemIndex{
		Chain:   targetDataIndex.Chain,
		Address: targetDataIndex.Address,
		TokenID: native.TokenID,
	}, nil

}

func (f *nftOperator) tryGetOnChainItem(ctx context.Context, from *tokentypes.OnChainItemIndex, toChain string) (*tokentypes.OnChainItemIndex, error) {
	toOnChainItemResp, err := tokentypes.NewQueryClient(f.rarimo).OnChainItemByOther(ctx, &tokentypes.QueryGetOnChainItemByOtherRequest{
		Chain:       from.Chain,
		Address:     from.Address,
		TokenID:     from.TokenID,
		TargetChain: toChain,
	})

	if err != nil {
		res, ok := status.FromError(err)
		if ok && res.Code() == codes.NotFound {
			return nil, nil
		}

		return nil, err
	}

	return toOnChainItemResp.Item.Index, nil
}

func (f *nftOperator) getNativeData(ctx context.Context, from *tokentypes.OnChainItemIndex) (*tokentypes.CollectionData, error) {
	collectionResp, err := tokentypes.NewQueryClient(f.rarimo).CollectionByCollectionData(ctx, &tokentypes.QueryGetCollectionByCollectionDataRequest{Chain: from.Chain, Address: from.Address})
	if err != nil {
		return nil, errors.Wrap(err, "error fetching collection")
	}

	nativeCollectionData, err := tokentypes.NewQueryClient(f.rarimo).NativeCollectionData(ctx, &tokentypes.QueryGetNativeCollectionDataRequest{Collection: collectionResp.Collection.Index})
	if err != nil {
		return nil, errors.Wrap(err, "error fetching native collection data")
	}

	return &nativeCollectionData.Data, nil
}

func (f *nftOperator) getTargetDataIndex(ctx context.Context, from *tokentypes.OnChainItemIndex, targetChain string) (*tokentypes.CollectionDataIndex, error) {
	collectionResp, err := tokentypes.NewQueryClient(f.rarimo).CollectionByCollectionData(ctx, &tokentypes.QueryGetCollectionByCollectionDataRequest{Chain: from.Chain, Address: from.Address})
	if err != nil {
		return nil, errors.Wrap(err, "error fetching collection")
	}

	for _, index := range collectionResp.Collection.Data {
		if index.Chain == targetChain {
			return index, nil
		}
	}

	return nil, verifiers.ErrWrongOperationContent
}

func (f *nftOperator) getItemMeta(ctx context.Context, from *tokentypes.OnChainItemIndex) (*tokentypes.ItemMetadata, error) {
	// return empty meta if should not been provided
	_, err := tokentypes.NewQueryClient(f.rarimo).OnChainItem(ctx, &tokentypes.QueryGetOnChainItemRequest{Chain: from.Chain, Address: from.Address, TokenID: from.TokenID})
	if err == nil {
		return nil, nil
	}

	metadata, err := f.getMetadata(solana.MustPublicKeyFromBase58(from.TokenID))
	if err != nil {
		return nil, errors.Wrap(err, "error fetching metadata from chain")
	}

	imageUrl, imageHash, err := verifiers.GetImage(metadata.Data.URI)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching image")
	}

	return &tokentypes.ItemMetadata{
		ImageUri:  imageUrl,
		ImageHash: imageHash,
		Seed:      "", // Empty because we are creating item on Solana chain => Solana is a native chain for that token.
		Uri:       metadata.Data.URI,
	}, nil
}

func (f *nftOperator) getMetadata(mint solana.PublicKey) (*metaplex.Metadata, error) {
	metadataAddress, _, err := solana.FindTokenMetadataAddress(mint)
	if err != nil {
		return nil, errors.Wrap(err, "error generating metadata key")
	}

	metadataInfo, err := f.solana.GetAccountInfo(context.TODO(), metadataAddress)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching metadata account info")
	}

	metadata := new(metaplex.Metadata)
	return metadata, borsh.Deserialize(metadata, metadataInfo.Value.Data.GetBinary())
}

func (f *nftOperator) getTokenCollectionAddress(mint solana.PublicKey) (string, error) {
	metadata, err := f.getMetadata(mint)
	if err != nil {
		return "", errors.Wrap(err, "error getting metadata")
	}

	// TODO maybe check for zero collection address instead of verified value
	if metadata.Collection == nil || !metadata.Collection.Verified {
		return "", nil
	}

	return hexutil.Encode(metadata.Collection.Address.Bytes()), nil
}
