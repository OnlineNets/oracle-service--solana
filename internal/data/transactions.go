package data

const (
	TransactionsTableName           = "transactions"
	TransactionsHashColumn          = "hash"
	TransactionsTokenMintColumn     = "token_mint"
	TransactionsCollectionColumn    = "token_collection"
	TransactionsTokenIdColumn       = "token_id"
	TransactionsTargetNetworkColumn = "target_network"
	TransactionsReceiverColumn      = "receiver"
	TransactionsTokenTypeColumn     = "token_type"
)

type Transaction struct {
	Id            uint64 `db:"id" structs:"-"`
	Hash          string `db:"hash" structs:"hash"`
	TokenMint     string `db:"token_mint" structs:"token_mint"`
	Collection    string `db:"collection" structs:"collection"`
	TokenId       string `db:"token_id" structs:"token_id"`
	TargetNetwork string `db:"target_network" structs:"target_network"`
	Receiver      string `db:"receiver" structs:"receiver"`
	TokenType     int16  `db:"token_type" structs:"token_type"`
}
