package voter

import (
	"math/big"

	"github.com/olegfomenko/solana-go"
)

func CastAmount(currentAmount string, currentDecimals uint8, targetDecimals uint8) string {
	if currentDecimals == targetDecimals {
		return currentAmount
	}

	value, _ := new(big.Int).SetString(currentAmount, 10)

	if currentDecimals < targetDecimals {
		for i := uint8(0); i < targetDecimals-currentDecimals; i++ {
			value.Mul(value, new(big.Int).SetInt64(10))
		}

		return value.String()
	}

	for i := uint8(0); i < currentDecimals-targetDecimals; i++ {
		value.Div(value, new(big.Int).SetInt64(10))
	}

	return value.String()
}

func getInstructionAccounts(accounts []solana.PublicKey, indexes []uint16) []solana.PublicKey {
	result := make([]solana.PublicKey, 0, len(indexes))
	for _, i := range indexes {
		result = append(result, accounts[i])
	}
	return result
}
