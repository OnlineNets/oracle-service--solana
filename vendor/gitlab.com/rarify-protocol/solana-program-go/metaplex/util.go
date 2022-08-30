package metaplex

import "github.com/olegfomenko/solana-go"

const (
	MetadataPrefix = "metadata"
	EditionPrefix  = "edition"
)

func FindTokenMasterEditionAddress(mint solana.PublicKey) (solana.PublicKey, uint8, error) {
	seed := [][]byte{
		[]byte(MetadataPrefix),
		solana.TokenMetadataProgramID[:],
		mint[:],
		[]byte(EditionPrefix),
	}
	return solana.FindProgramAddress(seed, solana.TokenMetadataProgramID)
}
