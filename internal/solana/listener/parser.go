package listener

import (
	"github.com/near/borsh-go"
	"github.com/olegfomenko/solana-go"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarify-protocol/solana-proxy-svc/internal/solana/contract"
)

func (l *listener) parseWithdrawMetaplex(tx solana.Signature, accounts []solana.PublicKey, instruction solana.CompiledInstruction) error {
	l.log.Info("Found new withdraw in tx: " + tx.String())
	var instructionData contract.WithdrawMetaplex

	err := borsh.Deserialize(&instructionData, instruction.Data)
	if err != nil {
		return errors.Wrap(err, "error deserializing instruction data")
	}

	// TODO save to db ow send webhook
	// we can get account field using
	// accounts[instruction.Accounts[WithdrawBridgeAdminIndex]].String(),

	return nil
}
