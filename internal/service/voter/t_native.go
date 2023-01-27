package voter

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gogo/protobuf/proto"
	"github.com/near/borsh-go"
	"github.com/olegfomenko/solana-go"
	"gitlab.com/distributed_lab/logan/v3/errors"
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

func NewNativeOperator(chain string, rarimo *grpc.ClientConn) *nativeOperator {
	return &nativeOperator{
		chain:  chain,
		rarimo: rarimo,
	}
}

func (n *nativeOperator) ParseTransaction(ctx context.Context, accounts []solana.PublicKey, instruction solana.CompiledInstruction, transfer *rarimotypes.Transfer) error {
	msg, err := n.GetMessage(ctx, accounts, instruction)
	if err != nil {
		return errors.Wrap(err, "error getting message")
	}

	msg.Tx = transfer.Tx
	msg.EventId = transfer.EventId

	transferResp, err := rarimotypes.NewQueryClient(n.rarimo).Transfer(ctx, &rarimotypes.QueryGetTransferRequest{Msg: *msg})
	if err != nil {
		return errors.Wrap(err, "error querying transfer from core")
	}

	if proto.Equal(&transferResp.Transfer, transfer) {
		return verifiers.ErrWrongOperationContent
	}

	return nil
}

func (n *nativeOperator) GetMessage(ctx context.Context, _ []solana.PublicKey, instruction solana.CompiledInstruction) (*rarimotypes.MsgCreateTransferOp, error) {
	var args contract.DepositNativeArgs
	if err := borsh.Deserialize(&args, instruction.Data); err != nil {
		return nil, errors.Wrap(err, "error desser tx args")
	}

	from := &tokentypes.OnChainItemIndex{
		Chain:   n.chain,
		Address: "",
		TokenID: "",
	}

	to, err := n.getTo(ctx, args.NetworkTo)
	if err != nil {
		return nil, err
	}

	msg := &rarimotypes.MsgCreateTransferOp{
		Receiver: args.ReceiverAddress,
		Amount:   fmt.Sprint(args.Amount),
		From:     from,
		To:       to,
	}

	if args.BundleData != nil && len(*args.BundleData) > 0 && args.BundleSeed != nil {
		msg.BundleData = hexutil.Encode(*args.BundleData)
		msg.BundleSalt = hexutil.Encode((*args.BundleSeed)[:])
	}

	return msg, nil
}

func (n *nativeOperator) getTo(ctx context.Context, chain string) (*tokentypes.OnChainItemIndex, error) {
	fromOnChainResp, err := tokentypes.NewQueryClient(n.rarimo).OnChainItem(ctx, &tokentypes.QueryGetOnChainItemRequest{Chain: n.chain})
	if err != nil {
		return nil, errors.Wrap(err, "error fetching on chain item")
	}

	itemResp, err := tokentypes.NewQueryClient(n.rarimo).Item(ctx, &tokentypes.QueryGetItemRequest{Index: fromOnChainResp.Item.Item})
	if err != nil {
		return nil, errors.Wrap(err, "error fetching item")
	}

	for _, index := range itemResp.Item.OnChain {
		if index.Chain == chain {
			return index, nil
		}
	}

	return nil, verifiers.ErrWrongOperationContent
}
