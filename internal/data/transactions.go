package data

const (
	TransactionsTableName           = "transactions"
	TransactionsHashColumn          = "hash"
	TransactionsTokenAddressColumn  = "token_address"
	TransactionsTokenIdColumn       = "token_id"
	TransactionsTargetNetworkColumn = "target_network"
	TransactionsReceiverColumn      = "receiver"
)

type Transactions struct {
	Hash          string `db:"hash" structs:"-"`
	TokenAddress  string `db:"token_address" structs:"-"`
	TokenId       string `db:"token_id" structs:"-"`
	TargetNetwork string `db:"target_network" structs:"-"`
	Receiver      string `db:"receiver" structs:"-"`
}
