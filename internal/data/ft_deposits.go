package data

type FTDeposit struct {
	Id            uint64 `db:"id" structs:"-"`
	Hash          string `db:"hash" structs:"hash"`
	InstructionId uint32 `db:"instruction_id" structs:"instruction_id"`
	Sender        string `db:"sender" structs:"sender"`
	Receiver      string `db:"receiver" structs:"receiver"`
	TargetNetwork string `db:"target_network" structs:"target_network"`
	Amount        uint64 `db:"amount" structs:"amount"`
	Mint          string `db:"mint" structs:"mint"`
}
