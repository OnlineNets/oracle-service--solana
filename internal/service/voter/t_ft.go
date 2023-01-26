package voter

import (
	"context"
	"fmt"

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

type ftOperator struct {
	chain  string
	rarimo *grpc.ClientConn
}

func (f *ftOperator) ParseTransaction(ctx context.Context, accounts []solana.PublicKey, instruction solana.CompiledInstruction, transfer *rarimotypes.Transfer) error {
	msg, err := f.GetMessage(ctx, accounts, instruction)
	if err != nil {
		return err
	}

	msg.Tx = transfer.Tx
	msg.EventId = transfer.EventId

	transferResp, err := rarimotypes.NewQueryClient(f.rarimo).Transfer(ctx, &rarimotypes.QueryGetTransferRequest{Msg: *msg})
	if err != nil {
		return err
	}

	if proto.Equal(&transferResp.Transfer, transfer) {
		return verifiers.ErrWrongOperationContent
	}

	return nil
}

func (f *ftOperator) GetMessage(ctx context.Context, accounts []solana.PublicKey, instruction solana.CompiledInstruction) (*rarimotypes.MsgCreateTransferOp, error) {
	var args contract.DepositFTArgs
	if err := borsh.Deserialize(&args, instruction.Data); err != nil {
		return nil, err
	}

	address := hexutil.Encode(accounts[contract.DepositFTMintIndex].Bytes())

	from := &tokentypes.OnChainItemIndex{
		Chain:   f.chain,
		Address: address,
		TokenID: address,
	}

	to, err := f.getTo(ctx, from, args.NetworkTo)
	if to == nil {
		return nil, err
	}

	var bundleData, bundleSeed string
	if args.BundleData != nil && len(*args.BundleData) > 0 && args.BundleSeed != nil {
		bundleData = hexutil.Encode(*args.BundleData)
		bundleSeed = hexutil.Encode((*args.BundleSeed)[:])
	}

	return &rarimotypes.MsgCreateTransferOp{
		Receiver:   args.ReceiverAddress,
		Amount:     fmt.Sprint(args.Amount),
		BundleData: bundleData,
		BundleSalt: bundleSeed,
		From:       from,
		To:         to,
	}, nil
}

func (f *ftOperator) getTo(ctx context.Context, from *tokentypes.OnChainItemIndex, chain string) (*tokentypes.OnChainItemIndex, error) {
	fromOnChainResp, err := tokentypes.NewQueryClient(f.rarimo).OnChainItem(ctx, &tokentypes.QueryGetOnChainItemRequest{Chain: f.chain, Address: from.Address})
	if err != nil {
		return nil, err
	}

	itemResp, err := tokentypes.NewQueryClient(f.rarimo).Item(ctx, &tokentypes.QueryGetItemRequest{Index: fromOnChainResp.Item.Item})
	if err != nil {
		return nil, err
	}

	for _, index := range itemResp.Item.OnChain {
		if index.Chain == chain {
			return index, nil
		}
	}

	return nil, verifiers.ErrWrongOperationContent
}
