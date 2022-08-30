package contract

import (
	"fmt"

	"github.com/near/borsh-go"
	"github.com/olegfomenko/solana-go"
	"gitlab.com/rarify-protocol/solana-program-go/metaplex"
)

type Instruction borsh.Enum

const (
	InstructionInitAdmin Instruction = iota
	InstructionTransferOwnership
	InstructionDepositNative
	InstructionDepositFT
	InstructionDepositNFT
	InstructionWithdrawNative
	InstructionWithdrawFT
	InstructionWithdrawNFT
	InstructionMintFT
	InstructionMintNFT
)

type TokenType borsh.Enum

const (
	Native TokenType = iota
	ERC20
	ERC721
	ERC1155
	MetaplexFT
	MetaplexNFT
)

const (
	InitializeAdminBridgeAdminIndex = iota
	InitializeAdminFeePayerIndex
)

type InitializeAdminArgs struct {
	Instruction Instruction
	PublicKey   [64]byte
	Seeds       [32]byte
}

const (
	TransferOwnershipBridgeAdminIndex = iota
)

type TransferOwnershipArgs struct {
	Instruction  Instruction
	NewPublicKey [64]byte
	Signature    [64]byte
	RecoveryId   byte
	Seeds        [32]byte
}

const (
	DepositNativeBridgeAdminIndex = iota
	DepositNativeDepositIndex
	DepositNativeOwnerIndex
)

type DepositNativeArgs struct {
	Instruction     Instruction
	Amount          uint64
	NetworkTo       string
	ReceiverAddress string
	Seeds           [32]byte
	Nonce           [32]byte
}

const (
	DepositFTBridgeAdminIndex = iota
	DepositFTMintIndex
	DepositFTOwnerAssocIndex
	DepositFTBridgeAssocIndex
	DepositFTDepositIndex
	DepositFTOwnerIndex
)

type DepositFTArgs struct {
	Instruction     Instruction
	Amount          uint64
	NetworkTo       string
	ReceiverAddress string
	Seeds           [32]byte
	Nonce           [32]byte
}

const (
	DepositNFTBridgeAdminIndex = iota
	DepositNFTMintIndex
	DepositNFTOwnerAssocIndex
	DepositNFTBridgeAssocIndex
	DepositNFTDepositIndex
	DepositNFTOwnerIndex
)

type DepositNFTArgs struct {
	Instruction     Instruction
	NetworkTo       string
	ReceiverAddress string
	Seeds           [32]byte
	Nonce           [32]byte
}

type SignedContent struct {
	TxHash      string
	AddressFrom string
	TokenIdFrom string
	NetworkFrom string
	Amount      uint64
	TokenType   TokenType
}

const (
	WithdrawNativeBridgeAdminIndex = iota
	WithdrawNativeOwnerIndex
	WithdrawNativeWithdrawIndex
)

const (
	WithdrawFTBridgeAdminIndex = iota
	WithdrawFTMintIndex
	WithdrawFTOwnerIndex
	WithdrawFTOwnerAssocIndex
	WithdrawFTBridgeAssocIndex
	WithdrawFTWithdrawIndex
)

const (
	WithdrawNFTBridgeAdminIndex = iota
	WithdrawNFTMintIndex
	WithdrawNFTMetadataIndex
	WithdrawNFTOwnerIndex
	WithdrawNFTOwnerAssocIndex
	WithdrawNFTBridgeAssocIndex
	WithdrawNFTWithdrawIndex
)

type WithdrawArgs struct {
	Instruction Instruction
	Content     SignedContent
	Signature   [64]byte
	RecoveryId  byte
	Path        [][32]byte
	Root        [32]byte
	Seeds       [32]byte
}

type MintFTArgs struct {
	Instruction Instruction
	Data        metaplex.DataV2
	Seeds       [32]byte
	Amount      uint64
	Decimals    byte
}

type MintNFTArgs struct {
	Instruction Instruction
	Data        metaplex.DataV2
	Seeds       [32]byte
	Verify      bool
}

func InitializeAdminInstruction(programId, bridgeAdmin, feePayer solana.PublicKey, args InitializeAdminArgs) (solana.Instruction, error) {
	args.Instruction = InstructionInitAdmin

	accounts := solana.AccountMetaSlice(make([]*solana.AccountMeta, 0, 4))
	accounts.Append(solana.NewAccountMeta(bridgeAdmin, true, false))
	accounts.Append(solana.NewAccountMeta(feePayer, true, true))
	accounts.Append(solana.NewAccountMeta(solana.SystemProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SysVarRentPubkey, false, false))

	data, err := borsh.Serialize(args)
	if err != nil {
		return nil, err
	}

	return solana.NewInstruction(
		programId,
		accounts,
		data,
	), nil
}

func TransferOwnershipInstruction(programId, bridgeAdmin solana.PublicKey, args TransferOwnershipArgs) (solana.Instruction, error) {
	args.Instruction = InstructionTransferOwnership

	accounts := solana.AccountMetaSlice(make([]*solana.AccountMeta, 0, 1))
	accounts.Append(solana.NewAccountMeta(bridgeAdmin, true, false))

	data, err := borsh.Serialize(args)
	if err != nil {
		return nil, err
	}

	return solana.NewInstruction(
		programId,
		accounts,
		data,
	), nil
}

func DepositNativeInstruction(programId, bridgeAdmin, deposit, owner solana.PublicKey, args DepositNativeArgs) (solana.Instruction, error) {
	args.Instruction = InstructionDepositNative

	accounts := solana.AccountMetaSlice(make([]*solana.AccountMeta, 0, 5))
	accounts.Append(solana.NewAccountMeta(bridgeAdmin, true, false))
	accounts.Append(solana.NewAccountMeta(deposit, true, false))
	accounts.Append(solana.NewAccountMeta(owner, true, true))
	accounts.Append(solana.NewAccountMeta(solana.SystemProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SysVarRentPubkey, false, false))

	data, err := borsh.Serialize(args)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Deposit instruction data len: %d\n", len(data))

	return solana.NewInstruction(
		programId,
		accounts,
		data,
	), nil
}

func DepositFTInstruction(programId, bridgeAdmin, mint, deposit, owner solana.PublicKey, args DepositFTArgs) (solana.Instruction, error) {
	args.Instruction = InstructionDepositFT

	bridgeAssoc, _, err := solana.FindAssociatedTokenAddress(bridgeAdmin, mint)
	if err != nil {
		return nil, err
	}

	ownerAssoc, _, err := solana.FindAssociatedTokenAddress(owner, mint)
	if err != nil {
		return nil, err
	}

	accounts := solana.AccountMetaSlice(make([]*solana.AccountMeta, 0, 10))
	accounts.Append(solana.NewAccountMeta(bridgeAdmin, false, false))
	accounts.Append(solana.NewAccountMeta(mint, false, false))
	accounts.Append(solana.NewAccountMeta(ownerAssoc, true, false))
	accounts.Append(solana.NewAccountMeta(bridgeAssoc, true, false))
	accounts.Append(solana.NewAccountMeta(deposit, true, false))
	accounts.Append(solana.NewAccountMeta(owner, true, true))
	accounts.Append(solana.NewAccountMeta(solana.TokenProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SystemProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SysVarRentPubkey, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SPLAssociatedTokenAccountProgramID, false, false))

	data, err := borsh.Serialize(args)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Deposit instruction data len: %d\n", len(data))

	return solana.NewInstruction(
		programId,
		accounts,
		data,
	), nil
}

func DepositNFTInstruction(programId, bridgeAdmin, mint, deposit, owner solana.PublicKey, args DepositNFTArgs) (solana.Instruction, error) {
	args.Instruction = InstructionDepositNFT

	bridgeAssoc, _, err := solana.FindAssociatedTokenAddress(bridgeAdmin, mint)
	if err != nil {
		return nil, err
	}

	ownerAssoc, _, err := solana.FindAssociatedTokenAddress(owner, mint)
	if err != nil {
		return nil, err
	}

	accounts := solana.AccountMetaSlice(make([]*solana.AccountMeta, 0, 10))
	accounts.Append(solana.NewAccountMeta(bridgeAdmin, false, false))
	accounts.Append(solana.NewAccountMeta(mint, false, false))
	accounts.Append(solana.NewAccountMeta(ownerAssoc, true, false))
	accounts.Append(solana.NewAccountMeta(bridgeAssoc, true, false))
	accounts.Append(solana.NewAccountMeta(deposit, true, false))
	accounts.Append(solana.NewAccountMeta(owner, true, true))
	accounts.Append(solana.NewAccountMeta(solana.TokenProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SystemProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SysVarRentPubkey, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SPLAssociatedTokenAccountProgramID, false, false))

	data, err := borsh.Serialize(args)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Deposit instruction data len: %d\n", len(data))

	return solana.NewInstruction(
		programId,
		accounts,
		data,
	), nil
}

func WithdrawNativeInstruction(programId, bridgeAdmin, owner, withdraw solana.PublicKey, args WithdrawArgs) (solana.Instruction, error) {
	args.Instruction = InstructionWithdrawNative

	accounts := solana.AccountMetaSlice(make([]*solana.AccountMeta, 0, 5))
	accounts.Append(solana.NewAccountMeta(bridgeAdmin, true, false))
	accounts.Append(solana.NewAccountMeta(owner, true, true))
	accounts.Append(solana.NewAccountMeta(withdraw, true, false))
	accounts.Append(solana.NewAccountMeta(solana.SystemProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SysVarRentPubkey, false, false))

	data, err := borsh.Serialize(args)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Withdraw instruction data len: %d\n", len(data))

	return solana.NewInstruction(
		programId,
		accounts,
		data,
	), nil
}

func WithdrawFTInstruction(programId, bridgeAdmin, mint, owner, withdraw solana.PublicKey, args WithdrawArgs) (solana.Instruction, error) {
	args.Instruction = InstructionWithdrawFT

	bridgeAssoc, _, err := solana.FindAssociatedTokenAddress(bridgeAdmin, mint)
	if err != nil {
		return nil, err
	}

	ownerAssoc, _, err := solana.FindAssociatedTokenAddress(owner, mint)
	if err != nil {
		return nil, err
	}

	accounts := solana.AccountMetaSlice(make([]*solana.AccountMeta, 0, 10))
	accounts.Append(solana.NewAccountMeta(bridgeAdmin, false, false))
	accounts.Append(solana.NewAccountMeta(mint, false, false))
	accounts.Append(solana.NewAccountMeta(owner, true, true))
	accounts.Append(solana.NewAccountMeta(ownerAssoc, true, false))
	accounts.Append(solana.NewAccountMeta(bridgeAssoc, true, false))
	accounts.Append(solana.NewAccountMeta(withdraw, true, false))
	accounts.Append(solana.NewAccountMeta(solana.TokenProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SystemProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SysVarRentPubkey, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SPLAssociatedTokenAccountProgramID, false, false))

	data, err := borsh.Serialize(args)
	if err != nil {
		return nil, err
	}

	return solana.NewInstruction(
		programId,
		accounts,
		data,
	), nil
}

func WithdrawNFTInstruction(programId, bridgeAdmin, mint, owner, withdraw solana.PublicKey, args WithdrawArgs) (solana.Instruction, error) {
	args.Instruction = InstructionWithdrawNFT

	bridgeAssoc, _, err := solana.FindAssociatedTokenAddress(bridgeAdmin, mint)
	if err != nil {
		return nil, err
	}

	ownerAssoc, _, err := solana.FindAssociatedTokenAddress(owner, mint)
	if err != nil {
		return nil, err
	}

	metadata, _, err := solana.FindTokenMetadataAddress(mint)
	if err != nil {
		return nil, err
	}

	accounts := solana.AccountMetaSlice(make([]*solana.AccountMeta, 0, 11))
	accounts.Append(solana.NewAccountMeta(bridgeAdmin, false, false))
	accounts.Append(solana.NewAccountMeta(mint, false, false))
	accounts.Append(solana.NewAccountMeta(metadata, false, false))
	accounts.Append(solana.NewAccountMeta(owner, true, true))
	accounts.Append(solana.NewAccountMeta(ownerAssoc, true, false))
	accounts.Append(solana.NewAccountMeta(bridgeAssoc, true, false))
	accounts.Append(solana.NewAccountMeta(withdraw, true, false))
	accounts.Append(solana.NewAccountMeta(solana.TokenProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SystemProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SysVarRentPubkey, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SPLAssociatedTokenAccountProgramID, false, false))

	data, err := borsh.Serialize(args)
	if err != nil {
		return nil, err
	}

	return solana.NewInstruction(
		programId,
		accounts,
		data,
	), nil
}

func MintFTInstruction(programId, bridgeAdmin, mint, payer solana.PublicKey, args MintFTArgs) (solana.Instruction, error) {
	args.Instruction = InstructionMintFT

	metadata, _, err := solana.FindTokenMetadataAddress(mint)
	if err != nil {
		return nil, err
	}

	bridgeAssoc, _, err := solana.FindAssociatedTokenAddress(bridgeAdmin, mint)
	if err != nil {
		return nil, err
	}

	accounts := solana.AccountMetaSlice(make([]*solana.AccountMeta, 0, 10))
	accounts.Append(solana.NewAccountMeta(bridgeAdmin, true, false))
	accounts.Append(solana.NewAccountMeta(mint, true, false))
	accounts.Append(solana.NewAccountMeta(bridgeAssoc, true, false))
	accounts.Append(solana.NewAccountMeta(metadata, true, false))
	accounts.Append(solana.NewAccountMeta(payer, true, true))

	accounts.Append(solana.NewAccountMeta(solana.TokenProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.TokenMetadataProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SysVarRentPubkey, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SystemProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SPLAssociatedTokenAccountProgramID, false, false))

	data, err := borsh.Serialize(args)
	if err != nil {
		return nil, err
	}

	return solana.NewInstruction(
		programId,
		accounts,
		data,
	), nil
}

func MintNFTVerifiedInstruction(programId, bridgeAdmin, mint, payer, collection solana.PublicKey, args MintNFTArgs) (solana.Instruction, error) {
	args.Instruction = InstructionMintNFT
	args.Verify = true

	metadata, _, err := solana.FindTokenMetadataAddress(mint)
	if err != nil {
		return nil, err
	}

	masterEdition, _, err := metaplex.FindTokenMasterEditionAddress(mint)
	if err != nil {
		return nil, err
	}

	bridgeAssoc, _, err := solana.FindAssociatedTokenAddress(bridgeAdmin, mint)
	if err != nil {
		return nil, err
	}

	accounts := solana.AccountMetaSlice(make([]*solana.AccountMeta, 0, 11))
	accounts.Append(solana.NewAccountMeta(bridgeAdmin, true, false))
	accounts.Append(solana.NewAccountMeta(mint, true, false))
	accounts.Append(solana.NewAccountMeta(bridgeAssoc, true, false))
	accounts.Append(solana.NewAccountMeta(metadata, true, false))
	accounts.Append(solana.NewAccountMeta(masterEdition, true, false))

	accounts.Append(solana.NewAccountMeta(payer, true, true))

	accounts.Append(solana.NewAccountMeta(solana.TokenProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.TokenMetadataProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SysVarRentPubkey, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SystemProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SPLAssociatedTokenAccountProgramID, false, false))

	collectionMetadata, _, err := solana.FindTokenMetadataAddress(collection)
	if err != nil {
		return nil, err
	}

	collectionMasterEdition, _, err := metaplex.FindTokenMasterEditionAddress(collection)
	if err != nil {
		return nil, err
	}

	accounts.Append(solana.NewAccountMeta(collection, false, false))
	accounts.Append(solana.NewAccountMeta(collectionMetadata, false, false))
	accounts.Append(solana.NewAccountMeta(collectionMasterEdition, false, false))

	data, err := borsh.Serialize(args)
	if err != nil {
		return nil, err
	}

	return solana.NewInstruction(
		programId,
		accounts,
		data,
	), nil
}

func MintNFTInstruction(programId, bridgeAdmin, mint, payer solana.PublicKey, args MintNFTArgs) (solana.Instruction, error) {
	args.Instruction = InstructionMintNFT
	args.Verify = false

	metadata, _, err := solana.FindTokenMetadataAddress(mint)
	if err != nil {
		return nil, err
	}

	masterEdition, _, err := metaplex.FindTokenMasterEditionAddress(mint)
	if err != nil {
		return nil, err
	}

	bridgeAssoc, _, err := solana.FindAssociatedTokenAddress(bridgeAdmin, mint)
	if err != nil {
		return nil, err
	}

	accounts := solana.AccountMetaSlice(make([]*solana.AccountMeta, 0, 11))
	accounts.Append(solana.NewAccountMeta(bridgeAdmin, true, false))
	accounts.Append(solana.NewAccountMeta(mint, true, false))
	accounts.Append(solana.NewAccountMeta(bridgeAssoc, true, false))
	accounts.Append(solana.NewAccountMeta(metadata, true, false))
	accounts.Append(solana.NewAccountMeta(masterEdition, true, false))

	accounts.Append(solana.NewAccountMeta(payer, true, true))

	accounts.Append(solana.NewAccountMeta(solana.TokenProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.TokenMetadataProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SysVarRentPubkey, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SystemProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SPLAssociatedTokenAccountProgramID, false, false))

	data, err := borsh.Serialize(args)
	if err != nil {
		return nil, err
	}

	return solana.NewInstruction(
		programId,
		accounts,
		data,
	), nil
}
