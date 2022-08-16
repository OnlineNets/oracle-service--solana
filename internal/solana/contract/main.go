package contract

import (
	"fmt"

	"github.com/near/borsh-go"
	"github.com/olegfomenko/solana-go"
	"gitlab.com/rarify-protocol/solana-proxy-svc/internal/solana/metaplex"
)

type Instruction borsh.Enum

const (
	MaxAddressLen = 64
	MaxTokenIdLen = 64
	MaxNetworkLen = 64

	MaxNameLen   = 32
	MaxSymbolLen = 20
	MaxURILen    = 200

	RawWithdrawDataLen = 92
	RawMinDataLen      = 94

	MaxTxDataLen = 1232
)

const (
	InstructionInitAdmin Instruction = iota
	InstructionTransferOwnership
	InstructionDepositMetaplex
	InstructionWithdrawMetaplex
	InstructionMintMetaplex
)

const (
	InitAdminBridgeAdminIndex = 0

	TransferBridgeAdminIndex = 0

	DepositBridgeAdminIndex = 0
	DepositMintIndex        = 1
	DepositOwnerTokenIndex  = 2
	DepositBridgeTokenIndex = 3
	DepositDepositIndex     = 4
	DepositOwnerIndex       = 5

	WithdrawBridgeAdminIndex = 0
	WithdrawMintIndex        = 1
	WithdrawOwnerIndex       = 2
	WithdrawOwnerTokenIndex  = 3
	WithdrawBridgeTokenIndex = 4
	WithdrawWithdrawIndex    = 5
	WithdrawAdminIndex       = 6
)

type InitializeAdmin struct {
	Instruction Instruction
	Seeds       [32]byte
}

type TransferOwnership struct {
	Instruction Instruction
	NewAdmin    solana.PublicKey
	Seeds       [32]byte
}

type DepositMetaplex struct {
	Instruction     Instruction
	NetworkTo       string
	ReceiverAddress string
	Address         *string
	TokenId         *string
	Seeds           [32]byte
	Nonce           [32]byte
}

type WithdrawMetaplex struct {
	Instruction   Instruction
	DepositTx     string
	NetworkFrom   string
	SenderAddress string
	TokenId       *string
	Seeds         [32]byte
}

type MintMetaplex struct {
	Instruction Instruction
	Data        metaplex.DataV2
	Seeds       [32]byte
	Verified    bool
	TokenId     *string
	Address     *string
}

func InitializeAdminInstruction(programId, bridgeAdmin, admin solana.PublicKey, args InitializeAdmin) (solana.Instruction, error) {
	args.Instruction = InstructionInitAdmin

	accounts := solana.AccountMetaSlice(make([]*solana.AccountMeta, 0, 4))
	accounts.Append(solana.NewAccountMeta(bridgeAdmin, true, false))
	accounts.Append(solana.NewAccountMeta(admin, true, true))
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

func DepositMetaplexInstruction(programId, bridgeAdmin, mint, deposit, owner solana.PublicKey, args DepositMetaplex) (solana.Instruction, error) {
	args.Instruction = InstructionDepositMetaplex

	bridgeAssoc, _, err := solana.FindAssociatedTokenAddress(bridgeAdmin, mint)
	if err != nil {
		return nil, err
	}

	ownerAssoc, _, err := solana.FindAssociatedTokenAddress(owner, mint)
	if err != nil {
		return nil, err
	}

	accounts := solana.AccountMetaSlice(make([]*solana.AccountMeta, 0, 8))
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

func WithdrawMetaplexInstruction(programId, bridgeAdmin, mint, owner, withdraw, admin solana.PublicKey, args WithdrawMetaplex) (solana.Instruction, error) {
	args.Instruction = InstructionWithdrawMetaplex

	bridgeAssoc, _, err := solana.FindAssociatedTokenAddress(bridgeAdmin, mint)
	if err != nil {
		return nil, err
	}

	ownerAssoc, _, err := solana.FindAssociatedTokenAddress(owner, mint)
	if err != nil {
		return nil, err
	}

	accounts := solana.AccountMetaSlice(make([]*solana.AccountMeta, 0, 9))
	accounts.Append(solana.NewAccountMeta(bridgeAdmin, false, false))
	accounts.Append(solana.NewAccountMeta(mint, false, false))
	accounts.Append(solana.NewAccountMeta(owner, true, true))
	accounts.Append(solana.NewAccountMeta(ownerAssoc, true, false))
	accounts.Append(solana.NewAccountMeta(bridgeAssoc, true, false))
	accounts.Append(solana.NewAccountMeta(withdraw, true, false))
	accounts.Append(solana.NewAccountMeta(admin, false, true))
	accounts.Append(solana.NewAccountMeta(solana.TokenProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SystemProgramID, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SysVarRentPubkey, false, false))
	accounts.Append(solana.NewAccountMeta(solana.SPLAssociatedTokenAccountProgramID, false, false))

	data, err := borsh.Serialize(args)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Withdraw instruction data len: %d\n", len(data))

	fmt.Printf("Sz token_id: %d\n", len([]byte(*args.TokenId)))
	fmt.Printf("Sz network: %d\n", len([]byte(args.NetworkFrom)))
	fmt.Printf("Sz tx: %d\n", len([]byte(args.DepositTx)))
	fmt.Printf("Sz sender: %d\n", len([]byte(args.SenderAddress)))

	fmt.Printf("Sz data raw: %d\n", len(data)-len([]byte(args.SenderAddress))-len([]byte(args.NetworkFrom))-len([]byte(args.DepositTx))-len([]byte(*args.TokenId)))

	return solana.NewInstruction(
		programId,
		accounts,
		data,
	), nil
}

func MintVerifiedMetaplexInstruction(
	programId,
	bridgeAdmin,
	mint,
	admin,
	payer,
	collection solana.PublicKey,
	args MintMetaplex,
) (solana.Instruction, error) {
	args.Instruction = InstructionMintMetaplex
	args.Verified = true

	metadata, _, err := solana.FindTokenMetadataAddress(mint)
	if err != nil {
		return nil, err
	}

	masterEdition, _, err := FindTokenMasterEditionAddress(mint)
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

	accounts.Append(solana.NewAccountMeta(admin, false, true))
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

	collectionMasterEdition, _, err := FindTokenMasterEditionAddress(collection)
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

	fmt.Printf("Mint instruction data len: %d\n", len(data))

	fmt.Printf("Sz name: %d\n", len([]byte(args.Data.Name)))
	fmt.Printf("Sz uri: %d\n", len([]byte(args.Data.URI)))
	fmt.Printf("Sz symbol: %d\n", len([]byte(args.Data.Symbol)))

	fmt.Printf("Sz token_id: %d\n", len([]byte(*args.TokenId)))
	fmt.Printf("Sz address: %d\n", len([]byte(*args.Address)))

	fmt.Printf("Sz data raw: %d\n", len(data)-len([]byte(args.Data.Name))-len([]byte(args.Data.URI))-len([]byte(args.Data.Symbol))-len([]byte(*args.TokenId))-len([]byte(*args.Address)))

	return solana.NewInstruction(
		programId,
		accounts,
		data,
	), nil
}
