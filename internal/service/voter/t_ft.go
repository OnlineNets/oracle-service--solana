package voter

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gogo/protobuf/proto"
	"github.com/near/borsh-go"
	"github.com/olegfomenko/solana-go"
	oracletypes "github.com/rarimo/rarimo-core/x/oraclemanager/types"
	rarimotypes "github.com/rarimo/rarimo-core/x/rarimocore/types"
	tokentypes "github.com/rarimo/rarimo-core/x/tokenmanager/types"
	"github.com/rarimo/saver-grpc-lib/voter/verifiers"
	"github.com/rarimo/solana-program-go/contracts/bridge"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"google.golang.org/grpc"
)

type ftOperator struct {
	chain  string
	log    *logan.Entry
	rarimo *grpc.ClientConn
}

func NewFTOperator(chain string, log *logan.Entry, rarimo *grpc.ClientConn) *ftOperator {
	return &ftOperator{
		chain:  chain,
		log:    log,
		rarimo: rarimo,
	}
}

func (f *ftOperator) ParseTransaction(ctx context.Context, accounts []solana.PublicKey, instruction solana.CompiledInstruction, transfer *rarimotypes.Transfer) error {
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

func (f *ftOperator) GetMessage(ctx context.Context, accounts []solana.PublicKey, instruction solana.CompiledInstruction) (*oracletypes.MsgCreateTransferOp, error) {
	var args bridge.DepositFTArgs
	if err := borsh.Deserialize(&args, instruction.Data); err != nil {
		return nil, errors.Wrap(err, "error desser tx args")
	}

	address := hexutil.Encode(accounts[bridge.DepositFTMintIndex].Bytes())

	from := tokentypes.OnChainItemIndex{
		Chain:   f.chain,
		Address: address,
		TokenID: "",
	}

	to, err := f.getTo(ctx, &from, args.NetworkTo)
	if to == nil {
		return nil, err
	}

	msg := &oracletypes.MsgCreateTransferOp{
		Sender:   accounts[bridge.DepositFTOwnerIndex].String(),
		Receiver: args.ReceiverAddress,
		Amount:   fmt.Sprint(args.Amount),
		From:     from,
		To:       *to,
	}

	if args.BundleData != nil && len(*args.BundleData) > 0 && args.BundleSeed != nil {
		msg.BundleData = hexutil.Encode(*args.BundleData)
		msg.BundleSalt = hexutil.Encode((*args.BundleSeed)[:])
	}

	return msg, nil
}

func (f *ftOperator) getTo(ctx context.Context, from *tokentypes.OnChainItemIndex, chain string) (*tokentypes.OnChainItemIndex, error) {
	fromOnChainResp, err := tokentypes.NewQueryClient(f.rarimo).OnChainItem(ctx, &tokentypes.QueryGetOnChainItemRequest{Chain: f.chain, Address: from.Address})
	if err != nil {
		f.log.WithError(err).Error("error fetching on chain item")
		return nil, verifiers.ErrWrongOperationContent
	}

	itemResp, err := tokentypes.NewQueryClient(f.rarimo).Item(ctx, &tokentypes.QueryGetItemRequest{Index: fromOnChainResp.Item.Item})
	if err != nil {
		f.log.WithError(err).Error("error fetching item")
		return nil, verifiers.ErrWrongOperationContent
	}

	for _, index := range itemResp.Item.OnChain {
		if index.Chain == chain {
			return index, nil
		}
	}

	return nil, verifiers.ErrWrongOperationContent
}
