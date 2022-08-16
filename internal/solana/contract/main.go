package contract

import (
	"github.com/near/borsh-go"
)

type Instruction borsh.Enum

const (
	InstructionInitAdmin Instruction = iota
	InstructionTransferOwnership
	InstructionDepositMetaplex
	InstructionWithdrawMetaplex
	InstructionMintMetaplex
)

const (
	DepositBridgeAdminIndex = 0
	DepositMintIndex        = 1
	DepositOwnerTokenIndex  = 2
	DepositBridgeTokenIndex = 3
	DepositDepositIndex     = 4
	DepositOwnerIndex       = 5
)

type DepositMetaplex struct {
	Instruction     Instruction
	NetworkTo       string
	ReceiverAddress string
	Address         *string
	TokenId         *string
	Seeds           [32]byte
	Nonce           [32]byte
}
