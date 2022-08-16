package listener

import (
	"github.com/near/borsh-go"
	"github.com/olegfomenko/solana-go"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/data"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/solana/contract"
)

func (l *listener) parseDepositMetaplex(tx solana.Signature, accounts []solana.PublicKey, instruction solana.CompiledInstruction) error {
	l.log.Info("Found new deposit in tx: " + tx.String())
	var instructionData contract.DepositMetaplex

	err := borsh.Deserialize(&instructionData, instruction.Data)
	if err != nil {
		return errors.Wrap(err, "error deserializing instruction data")
	}

	_, err = l.Transactions().Create(data.Transaction{
		Hash:          tx.String(),
		TokenAddress:  accounts[instruction.Accounts[contract.DepositMintIndex]].String(),
		TokenId:       instructionData.TokenId,
		TargetNetwork: instructionData.NetworkTo,
		Receiver:      instructionData.Address,
	})

	return err
}
