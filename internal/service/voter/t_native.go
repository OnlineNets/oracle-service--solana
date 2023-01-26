package voter

import (
	"context"
	"strconv"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gogo/protobuf/proto"
	"github.com/near/borsh-go"
	"github.com/olegfomenko/solana-go"
	rarimotypes "gitlab.com/rarimo/rarimo-core/x/rarimocore/types"
	tokentypes "gitlab.com/rarimo/rarimo-core/x/tokenmanager/types"
	"gitlab.com/rarimo/savers/saver-grpc-lib/voter/verifiers"
	"gitlab.com/rarimo/solana-program-go/contract"
	"google.golang.org/grpc"
)

type nativeOperator struct {
	chain  string
	rarimo *grpc.ClientConn
}

func (n *nativeOperator) ParseTransaction(ctx context.Context, accounts []solana.PublicKey, instruction solana.CompiledInstruction, transfer *rarimotypes.Transfer) error {
	verifiedTransfer, err := n.GetOperation(ctx, accounts, instruction)
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

func (n *nativeOperator) GetOperation(ctx context.Context, _ []solana.PublicKey, instruction solana.CompiledInstruction) (*rarimotypes.Transfer, error) {
	var args contract.DepositNativeArgs
	if err := borsh.Deserialize(&args, instruction.Data); err != nil {
		return nil, err
	}

	fromOnChainResp, err := tokentypes.NewQueryClient(n.rarimo).OnChainItem(ctx, &tokentypes.QueryGetOnChainItemRequest{Chain: n.chain})
	if err != nil {
		return nil, err
	}

	itemResp, err := tokentypes.NewQueryClient(n.rarimo).Item(ctx, &tokentypes.QueryGetItemRequest{Index: fromOnChainResp.Item.Item})
	if err != nil {
		return nil, err
	}

	var from, to *tokentypes.OnChainItemIndex = fromOnChainResp.Item.Index, nil
	for _, index := range itemResp.Item.OnChain {
		if index.Chain == args.NetworkTo {
			to = index
			break
		}
	}

	if to == nil {
		return nil, verifiers.ErrWrongOperationContent
	}

	fromDataResp, err := tokentypes.NewQueryClient(n.rarimo).CollectionData(ctx, &tokentypes.QueryGetCollectionDataRequest{Chain: n.chain})
	if err != nil {
		return nil, err
	}

	toDataResp, err := tokentypes.NewQueryClient(n.rarimo).CollectionData(ctx, &tokentypes.QueryGetCollectionDataRequest{Chain: to.Chain, Address: to.Address})
	if err != nil {
		return nil, err
	}

	var bundleData, bundleSeed string
	if args.BundleData != nil && len(*args.BundleData) > 0 && args.BundleSeed != nil {
		bundleData = hexutil.Encode(*args.BundleData)
		bundleSeed = hexutil.Encode((*args.BundleSeed)[:])
	}

	return &rarimotypes.Transfer{
		Receiver:   args.ReceiverAddress,
		Amount:     CastAmount(strconv.FormatUint(args.Amount, 10), uint8(fromDataResp.Data.Decimals), uint8(toDataResp.Data.Decimals)),
		BundleData: bundleData,
		BundleSalt: bundleSeed,
		From:       from,
		To:         to,
	}, nil
}
