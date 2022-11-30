package data

//go:generate xo schema "postgres://solana:solana@localhost:5432/solana?sslmode=disable" -o ./ --single=schema.xo.go --src templates
//go:generate xo schema "postgres://solana:solana@localhost:5432/solana?sslmode=disable" -o pg --single=schema.xo.go --src=pg/templates --go-context=both
